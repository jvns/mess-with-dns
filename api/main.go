package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	_ "net/http/pprof"

	"github.com/honeycombio/honeycomb-opentelemetry-go"
	"github.com/honeycombio/otel-config-go/otelconfig"
	"github.com/jvns/mess-with-dns/records"
	"github.com/jvns/mess-with-dns/streamer"
	"github.com/jvns/mess-with-dns/users"
	"github.com/miekg/dns"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("main")

type Config struct {
	workdir           string // where to read ip2asn files and static files from
	requestDBFilename string
	userDBFilename    string
	// for cookies
	hashKey  string
	blockKey string
	// address of powerdns API
	powerdnsAddress string
	// where to listen for dnstap messages
	dnstapAddress string
}

func readConfig() (*Config, error) {
	workdir := os.Getenv("WORKDIR")
	if workdir == "" {
		workdir = "."
	}
	requestDBFilename := os.Getenv("REQUEST_DB_FILENAME")
	if requestDBFilename == "" {
		return nil, fmt.Errorf("REQUEST_DB_FILENAME must be set")
	}
	userDBFilename := os.Getenv("USER_DB_FILENAME")
	if userDBFilename == "" {
		return nil, fmt.Errorf("USER_DB_FILENAME must be set")
	}
	hashKey := os.Getenv("HASH_KEY")
	if hashKey == "" {
		return nil, fmt.Errorf("HASH_KEY must be set")
	}
	blockKey := os.Getenv("BLOCK_KEY")
	if blockKey == "" {
		return nil, fmt.Errorf("BLOCK_KEY must be set")
	}
	return &Config{
		workdir:           workdir,
		requestDBFilename: requestDBFilename,
		userDBFilename:    userDBFilename,
		hashKey:           hashKey,
		blockKey:          blockKey,
		powerdnsAddress:   "http://localhost:8081",
		dnstapAddress:     "localhost:7777",
	}, nil
}

func main() {
	// setup honeycomb
	bsp := honeycomb.NewBaggageSpanProcessor()

	otelShutdown, err := otelconfig.ConfigureOpenTelemetry(
		otelconfig.WithSpanProcessor(bsp),
	)
	if err != nil {
		log.Fatalf("error setting up OTel SDK - %e", err)
	}
	defer otelShutdown()
	// start pprof
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	config, err := readConfig()
	if err != nil {
		log.Fatalf("error reading config: %s", err)
	}
	ctx := context.Background()
	handler, err := createHandler(ctx, config)
	if err != nil {
		log.Fatalf(err.Error())
	}
	go handler.cleanup()

	port := ":53"
	if len(os.Args) > 1 {
		port = ":" + os.Args[1]
	}
	fmt.Println("Listening for UDP on port", port)
	go func() {
		srv := &dns.Server{Handler: handler, Addr: port, Net: "udp"}
		if err := srv.ListenAndServe(); err != nil {
			panic(fmt.Sprintf("Failed to set udp listener %s\n", err.Error()))
		}
	}()
	fmt.Println("Listening for TCP on port", port)
	go func() {
		srv := &dns.Server{Handler: handler, Addr: port, Net: "tcp"}
		if err := srv.ListenAndServe(); err != nil {
			panic(fmt.Sprintf("Failed to set tcp listener %s\n", err.Error()))
		}
	}()

	fmt.Println("Listening on :8080")
	err = (&http.Server{Addr: ":8080", Handler: createRoutes(handler)}).ListenAndServe()
	if err != nil {
		log.Fatalf("error starting server: %s", err.Error())
	}
}

func createHandler(ctx context.Context, config *Config) (*handler, error) {
	logger, err := streamer.Init(ctx, config.workdir, config.requestDBFilename, config.dnstapAddress)
	if err != nil {
		return nil, fmt.Errorf("error creating logger: %s", err.Error())
	}
	userService, err := users.Init(config.userDBFilename, config.hashKey, config.blockKey)
	if err != nil {
		return nil, fmt.Errorf("error connecting to user database: %s", err.Error())
	}
	rs := records.Init(config.powerdnsAddress, "not-a-secret")

	handler := &handler{
		rs:          rs,
		logger:      logger,
		userService: userService,
		workdir:     config.workdir,
	}
	return handler, nil
}

type handler struct {
	logger      *streamer.Logger
	rs          records.RecordService
	userService *users.UserService
	workdir     string
}

func returnError(w http.ResponseWriter, r *http.Request, err error, status int) {
	msg := fmt.Sprintf("Error [%d]: %s\n", status, err.Error())
	logMsg(r, msg)
	http.Error(w, err.Error(), status)
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
}

func logMsg(r *http.Request, msg string) {
	ip := r.Header.Get("X-Forwarded-For")
	ip = strings.Split(ip, ",")[0]
	fmt.Printf("[%s] %s\n", ip, msg)
}

func (handle *handler) serveDNS(w dns.ResponseWriter, r *dns.Msg) error {
	// Proxy it to localhost:5454
	c := &dns.Client{
		Net:         "udp",
		DialTimeout: time.Second * 1,
	}
	response, _, err := c.Exchange(r, "localhost:5454")

	if err != nil {
		return err
	}
	if response.Truncated {
		// try TCP instead if the response was truncated
		c.Net = "tcp"
		response, _, err = c.Exchange(r, "localhost:5454")
		if err != nil {
			return err
		}
	}
	err = w.WriteMsg(response)
	if err != nil {
		return err
	}
	err = handle.logger.Log(response, w)

	if err != nil {
		return err
	}
	return nil
}

// allow 50 QPS per IP
func servFail(r *dns.Msg) *dns.Msg {
	m := new(dns.Msg)
	m.SetReply(r)
	m.SetRcode(r, dns.RcodeServerFailure)
	return m
}

func (handle *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	ctx := context.Background()
	ip, _, _ := net.SplitHostPort(w.RemoteAddr().String())
	_, span := tracer.Start(ctx, "dns.request")
	if len(r.Question) > 0 {
		span.SetAttributes(attribute.String("dns.name", r.Question[0].Name))
		span.SetAttributes(attribute.String("dns.source", ip))
	}

	err := handle.serveDNS(w, r)

	if err != nil {
		// return a SERVFAIL
		fmt.Println("error serving DNS", err)
		span.RecordError(err)
		err = w.WriteMsg(servFail(r))
		if err != nil {
			fmt.Println("error writing SERVFAIL", err)
			span.RecordError(err)
		}
	}
	span.End()
}

func (handle *handler) cleanup() {
	ctx := context.Background()
	_, span := tracer.Start(ctx, "cleanup")
	defer span.End()
	for {
		fmt.Println("Deleting old records & requests...")
		err := handle.rs.DeleteOldRecords(ctx, time.Now())
		if err != nil {
			span.RecordError(fmt.Errorf("error deleting old records: %s", err))
			fmt.Println("error deleting old records:", err)
		}
		err = handle.logger.DeleteOldRequests(ctx)
		if err != nil {
			span.RecordError(fmt.Errorf("error deleting old requests: %s", err))
			fmt.Println("error deleting old requests:", err)
		}
		time.Sleep(time.Minute * 15)
	}
}
