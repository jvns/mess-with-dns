package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "net/http/pprof"

	"github.com/honeycombio/honeycomb-opentelemetry-go"
	"github.com/honeycombio/otel-config-go/otelconfig"
	"github.com/jvns/mess-with-dns/records"
	"github.com/jvns/mess-with-dns/streamer"
	"github.com/jvns/mess-with-dns/users"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

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
	logMsg(r, fmt.Sprintf("%s %s", r.Method, r.URL.Path))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")
	username, _ := handle.userService.ReadSessionUsername(r)
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
		// TODO: don't let people delete other people's records
		deleteRecord(ctx, username, p[1], rs, w, r)
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
	// GET /requeststream
	case r.Method == "GET" && p[0] == "requeststream":
		if !requireLogin(username, r.URL.Path, r, w) {
			return
		}
		streamRequests(ctx, handle.logger, username, w, r)
	// POST /login
	case r.Method == "GET" && n == 1 && p[0] == "login":
		w.Header().Set("Cache-Control", "no-store")
		loginRandom(handle.userService, w, r)
	default:
		// serve static files
		w.Header().Set("Cache-Control", "public, max-age=120")
		// get workdir
		http.ServeFile(w, r, handle.workdir+"/frontend/"+r.URL.Path)
	}
}
