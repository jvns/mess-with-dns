package main

import (
	"context"
	"fmt"
	"log"
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
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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
	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

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

	wrappedHandler := otelhttp.NewHandler(handler, "mess-with-dns-api")
	fmt.Println("Listening on :8080")
	err = (&http.Server{Addr: ":8080", Handler: wrappedHandler}).ListenAndServe()
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

func requireLogin(username string, page string, r *http.Request, w http.ResponseWriter) bool {
	w.Header().Set("Cache-Control", "no-store")
	if username == "" {
		returnError(w, r, fmt.Errorf("you must be logged in to access this page: %s", page), http.StatusUnauthorized)
		return false
	}
	return true
}

func (handle *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("http.path", r.URL.Path))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")
	username, _ := handle.userService.ReadSessionUsername(r)
	logMsg(r, fmt.Sprintf("%s %s (%s)", r.Method, r.URL.Path, username))
	span.SetAttributes(attribute.String("username", username))

	rs := handle.rs
	p := strings.Split(r.URL.Path, "/")[1:]
	n := len(p)
	switch {
	case p[0] == "health":
		healthCheck(w, r)
	// check host header for messwithdns.com
	case r.Host == "messwithdns.com" || r.Host == "www.messwithdns.com":
		// redirect to .net
		http.Redirect(w, r, "https://messwithdns.net"+r.URL.Path, http.StatusFound)
	// GET /records: get everything from USERNAME.messwithdns.com.
	case r.Method == "GET" && p[0] == "records":
		if !requireLogin(username, r.URL.Path, r, w) {
			return
		}
		getRecords(ctx, username, rs, w, r)
	// DELETE /records/<ID>:
	case r.Method == "DELETE" && n == 2 && p[0] == "records":
		if !requireLogin(username, r.URL.Path, r, w) {
			return
		}
		deleteRecord(ctx, username, p[1], rs, w, r)
	// DELETE /records:
	case r.Method == "DELETE" && n == 1 && p[0] == "records":
		if !requireLogin(username, r.URL.Path, r, w) {
			return
		}
		deleteAllRecords(ctx, username, rs, w, r)
	// POST /records/<ID>: updates a record
	case r.Method == "POST" && n == 2 && p[0] == "records" && len(p[1]) > 0:
		if !requireLogin(username, r.URL.Path, r, w) {
			return
		}
		updateRecord(ctx, username, p[1], rs, w, r)
	// POST /records/: add a new record
	case r.Method == "POST" && p[0] == "records":
		if !requireLogin(username, r.URL.Path, r, w) {
			return
		}
		createRecord(ctx, username, rs, w, r)
	// GET /requests
	case r.Method == "GET" && p[0] == "requests":
		if !requireLogin(username, r.URL.Path, r, w) {
			return
		}
		getRequests(ctx, handle.logger, username, w, r)
	// DELETE /requests
	case r.Method == "DELETE" && p[0] == "requests":
		if !requireLogin(username, r.URL.Path, r, w) {
			return
		}
		deleteRequests(ctx, handle.logger, username, w, r)
	// GET /requeststream/$USERNAME
	case r.Method == "GET" && p[0] == "requeststream":
		// Try setting the username based on the path if the cookie method
		// didn't work. Trying this because we were getting a lot of requests
		// for /requeststream with no username for some unknown reason
		if username == "" && len(p) >= 2 {
			username = p[1]
		}
		if !requireLogin(username, r.URL.Path, r, w) {
			return
		}
		streamRequests(ctx, handle.logger, username, w, r)
	// POST /login
	case r.Method == "GET" && n == 1 && p[0] == "login":
		w.Header().Set("Cache-Control", "no-store")
		loginRandom(handle.userService, handle.rs, w, r)
	default:
		// serve static files
		w.Header().Set("Cache-Control", "public, max-age=120")
		// get workdir
		http.ServeFile(w, r, handle.workdir+"/frontend/"+r.URL.Path)
	}
}

func (handle *handler) serveDNS(w dns.ResponseWriter, r *dns.Msg) error {
	// Proxy it to localhost:5454, using UDP or TCP depending on how the
	// request came in
	udp_or_tcp := w.RemoteAddr().Network()
	c := &dns.Client{Net: udp_or_tcp}
	c.DialTimeout = time.Second * 1
	response, _, err := c.Exchange(r, "localhost:5454")
	if err != nil {
		return err
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

func (handle *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	ctx := context.Background()
	_, span := tracer.Start(ctx, "dns.request")
	if len(r.Question) > 0 {
		span.SetAttributes(attribute.String("dns.name", r.Question[0].Name))
	}

	err := handle.serveDNS(w, r)

	if err != nil {
		// return a SERVFAIL
		fmt.Println("error serving DNS", err)
		span.RecordError(err)
		m := new(dns.Msg)
		m.SetReply(r)
		m.SetRcode(r, dns.RcodeServerFailure)
		err = w.WriteMsg(m)
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
