package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "net/http/pprof"

	"github.com/getsentry/sentry-go"
	"github.com/gorilla/websocket"
	"github.com/honeycombio/honeycomb-opentelemetry-go"
	"github.com/honeycombio/otel-config-go/otelconfig"
	"github.com/miekg/dns"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = otel.Tracer("main")

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
	if env := os.Getenv("SENTRY_DSN"); env != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn: env,
		})
		if err != nil {
			log.Fatalf("sentry.Init: %s", err)
		}
	}
	db, err := connect()
	if err != nil {
		panic(fmt.Sprintf("Error connecting to database: %s", err.Error()))
	}
	err = createTables(db)
	if err != nil {
		panic(fmt.Sprintf("Error creating tables: %s", err.Error()))
	}
	go cleanup(db)
	ctx := context.Background()
	soaSerial, err = GetSerial(ctx, db)
	if err != nil {
		panic(fmt.Sprintf("Error getting SOA serial: %s", err.Error()))
	}
	defer db.Close()
	ranges, err := ReadRanges()
	if err != nil {
		panic(fmt.Sprintf("Error reading ranges: %s", err.Error()))
	}
	handler := &handler{db: db, ipRanges: &ranges}
	// udp port command line argument
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
	go func() {
		srv := &dns.Server{Handler: handler, Addr: port, Net: "tcp"}
		if err := srv.ListenAndServe(); err != nil {
			panic(fmt.Sprintf("Failed to set tcp listener %s\n", err.Error()))
		}
	}()

	fmt.Println("Listening on :8080")

	wrappedHandler := otelhttp.NewHandler(handler, "mess-with-dns-api")
	err = (&http.Server{Addr: ":8080", Handler: wrappedHandler}).ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("Failed to start http listener %s\n", err.Error()))
	}
}

type handler struct {
	db       *sql.DB
	ipRanges *Ranges
}

var soaSerial uint32

func makeDomain(name string) string {
	return name + ".messwithdns.com."
}

func returnError(w http.ResponseWriter, r *http.Request, err error, status int) {
	msg := fmt.Sprintf("Error [%d]: %s\n", status, err.Error())
	logMsg(r, msg)
	sentry.CaptureException(err)
	http.Error(w, err.Error(), status)
}

func deleteRequests(ctx context.Context, db *sql.DB, name string, w http.ResponseWriter, r *http.Request) {
	err := DeleteRequestsForDomain(ctx, db, name)
	if err != nil {
		err := fmt.Errorf("Error deleting requests: %s", err.Error())
		returnError(w, r, err, http.StatusInternalServerError)
		return
	}
}

func getRequests(ctx context.Context, db *sql.DB, username string, w http.ResponseWriter, r *http.Request) {
	requests, err := GetRequests(ctx, db, username)
	if err != nil {
		err := fmt.Errorf("Error getting requests: %s", err.Error())
		returnError(w, r, err, http.StatusInternalServerError)
		return
	}
	jsonOutput, err := json.Marshal(requests)
	if err != nil {
		err := fmt.Errorf("Error marshalling json: %s", err.Error())
		returnError(w, r, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonOutput)
}

func streamRequests(ctx context.Context, db *sql.DB, subdomain string, w http.ResponseWriter, r *http.Request) {
	// create websocket connection
	conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	defer conn.Close()
	if err != nil {
		err := fmt.Errorf("Error creating websocket connection: %s", err.Error())
		returnError(w, r, err, http.StatusInternalServerError)
		return
	}
	logMsg(r, fmt.Sprintf("creating stream for %s", subdomain))
	stream := CreateStream(subdomain)
	defer stream.Delete()
	c := stream.Get()
	// I don't really understand this ping/pong stuff but it's what the gorilla docs say to do
	ticker := time.NewTicker(15 * time.Second)
	pongWait := time.Second * 60
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		select {
		case <-ticker.C:
			err := conn.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				fmt.Println("Error writing ping:", err)
				return
			}
		case msg := <-c:
			err := conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				fmt.Println("Error writing message:", err)
				return
			}
		}
	}
}

func createRecord(ctx context.Context, db *sql.DB, username string, w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err := fmt.Errorf("Error reading body: %s", err.Error())
		returnError(w, r, err, http.StatusInternalServerError)
		return
	}
	rr, err := ParseRecord(body)
	if err != nil {
		if strings.Contains(err.Error(), "Invalid RR") {
			returnError(w, r, fmt.Errorf("Oops, invalid domain name"), http.StatusBadRequest)
		} else {
			returnError(w, r, fmt.Errorf("Error parsing record: %s", err.Error()), http.StatusInternalServerError)
		}
		return
	}
	if err = validateDomainName(rr.Header().Name, username); err != nil {
		returnError(w, r, err, http.StatusBadRequest)
		return
	}
	InsertRecord(ctx, db, rr)
}

func deleteRecord(ctx context.Context, db *sql.DB, id string, w http.ResponseWriter, r *http.Request) {
	// parse int from id
	idInt, err := strconv.Atoi(id)
	if err != nil {
		err := fmt.Errorf("Error parsing id: %s", err.Error())
		returnError(w, r, err, http.StatusBadRequest)
		return
	}
	DeleteRecord(ctx, db, idInt)
}

func updateRecord(ctx context.Context, db *sql.DB, username string, id string, w http.ResponseWriter, r *http.Request) {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		returnError(w, r, fmt.Errorf("Error parsing id: %s", err.Error()), http.StatusBadRequest)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		returnError(w, r, fmt.Errorf("Error reading body: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	rr, err := ParseRecord(body)
	if err != nil {
		returnError(w, r, fmt.Errorf("Error parsing record: %s", err.Error()), http.StatusBadRequest)
		return
	}
	if err = validateDomainName(rr.Header().Name, username); err != nil {
		returnError(w, r, err, http.StatusBadRequest)
		return
	}
	UpdateRecord(ctx, db, idInt, rr)
}

func getDomains(ctx context.Context, db *sql.DB, username string, w http.ResponseWriter, r *http.Request) {
	records, err := GetRecordsForName(ctx, db, username)
	if err != nil {
		returnError(w, r, fmt.Errorf("Error getting records: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	jsonOutput, err := json.Marshal(records)
	if err != nil {
		returnError(w, r, fmt.Errorf("Error marshalling json: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonOutput)
}

func logMsg(r *http.Request, msg string) {
	ip := r.Header.Get("X-Forwarded-For")
	ip = strings.Split(ip, ",")[0]
	fmt.Printf("[%s] %s\n", ip, msg)
}

func requireLogin(username string, page string, r *http.Request, w http.ResponseWriter) bool {
	w.Header().Set("Cache-Control", "no-store")
	if username == "" {
		returnError(w, r, fmt.Errorf("You must be logged in to access this page: %s", page), http.StatusUnauthorized)
		return false
	}
	return true
}

func (handle *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logMsg(r, fmt.Sprintf("Request: %s", r.URL.Path))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")
	username, _ := ReadSessionUsername(r)

	p := strings.Split(r.URL.Path, "/")[1:]
	n := len(p)
	switch {
	case p[0] == "health":
		healthCheck(w, r)
	// check host header for messwithdns.com
	case r.Host == "messwithdns.com" || r.Host == "www.messwithdns.com":
		// redirect to .net
		http.Redirect(w, r, "https://messwithdns.net"+r.URL.Path, http.StatusFound)
	// GET /domain: get everything from USERNAME.messwithdns.com.
	case r.Method == "GET" && p[0] == "domains":
		if !requireLogin(username, r.URL.Path, r, w) {
			return
		}
		getDomains(ctx, handle.db, username, w, r)
	// GET /requests
	case r.Method == "GET" && p[0] == "requests":
		if !requireLogin(username, r.URL.Path, r, w) {
			return
		}
		getRequests(ctx, handle.db, username, w, r)
	// DELETE /requests
	case r.Method == "DELETE" && p[0] == "requests":
		if !requireLogin(username, r.URL.Path, r, w) {
			return
		}
		deleteRequests(ctx, handle.db, username, w, r)
	// GET /requeststream
	case r.Method == "GET" && p[0] == "requeststream":
		if !requireLogin(username, r.URL.Path, r, w) {
			return
		}
		streamRequests(ctx, handle.db, username, w, r)
	// POST /record/new: add a new record
	case r.Method == "POST" && n == 2 && p[0] == "record" && p[1] == "new":
		if !requireLogin(username, r.URL.Path, r, w) {
			return
		}
		createRecord(ctx, handle.db, username, w, r)
	// DELETE /record/<ID>:
	case r.Method == "DELETE" && n == 2 && p[0] == "record":
		if !requireLogin(username, r.URL.Path, r, w) {
			return
		}
		// TODO: don't let people delete other people's records
		deleteRecord(ctx, handle.db, p[1], w, r)
	// POST /record/<ID>: updates a record
	case r.Method == "POST" && n == 2 && p[0] == "record":
		if !requireLogin(username, r.URL.Path, r, w) {
			return
		}
		updateRecord(ctx, handle.db, username, p[1], w, r)
	// POST /login
	case r.Method == "GET" && n == 1 && p[0] == "login":
		w.Header().Set("Cache-Control", "no-store")
		loginRandom(handle.db, w, r)
	// GET /oauth-callback
	case r.Method == "GET" && p[0] == "oauth-callback":
		w.Header().Set("Cache-Control", "no-store")
		oauthCallback(w, r)
	default:
		// serve static files
		w.Header().Set("Cache-Control", "public, max-age=120")
		http.ServeFile(w, r, "./frontend/"+r.URL.Path)
	}
}

func (handle *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "dns.request")
	start := time.Now()
	fmt.Println("Received request: ", r.Question[0].String())
	msg := dnsResponse(ctx, handle.db, r, w)
	err := w.WriteMsg(msg)
	if err != nil {
		fmt.Println("Error writing response: ", err.Error())
		sentry.CaptureException(err)
	}
	span.End()

	ctx = context.Background()
	ctx, span = tracer.Start(ctx, "dns.request.log")
	defer span.End()

	// everything after this is just logging
	elapsed := time.Since(start)
	if len(msg.Answer) > 0 {
		fmt.Println("Response: ", msg.Answer[0].String(), elapsed)
	} else {
		// print elapsed time
		fmt.Println("Response: (no records found)", elapsed)

	}

	// check if it's a TCP address
	remote_addr := getIP(w)
	remote_host := lookupHost(ctx, handle.ipRanges, remote_addr)
	span.SetAttributes(attribute.String("dns.remote_addr", remote_addr.String()))
	span.SetAttributes(attribute.String("dns.remote_host", remote_host))
	span.SetAttributes(attribute.String("dns.question", r.Question[0].String()))
	span.SetAttributes(attribute.Int("dns.answer_count", len(msg.Answer)))
	err = LogRequest(ctx, handle.db, r, msg, remote_addr, remote_host)
	if err != nil {
		fmt.Println("Error logging request:", err)
		sentry.CaptureException(err)
	}
	fmt.Println("Logged request")
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	// make dns request to check if we're up
	m := new(dns.Msg)
	m.SetQuestion("glass99.messwithdns.com.", dns.TypeA)
	m.RecursionDesired = true
	c := new(dns.Client)
	c.DialTimeout = time.Second * 1
	//c.Net = "tcp"
	_, _, err := c.Exchange(m, "127.0.0.1:53")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error making DNS request: " + err.Error()))
		return
	}

	/*
		// get requests for glass99 and make sure it's a 200 ok
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://127.0.0.1:8080/requests", nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error creating request"))
			return
		}
		req.Header.Set("User-Agent", "healthcheck")
		req.Header.Set("Cookie", "session=MTY1MDEyMjU3MnxTcnB5M3ZvYmFKRXBhWXV0Y3kwWWNTTk5mU05Nb3hvRG5yajNkM2Fod2dNRVJ3MEJUX0RwTng2anduVGpOYVdTVENTdFY3aXNPWEJxVUNORXJlSGp8vaZ4BQTLfPwl6xy5VIvMQsqB2qiTjgss2RYWJUqCCTM=; username=glass99")
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != 200 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error requesting /requests"))
			return
		}
	*/
}

func getIP(w dns.ResponseWriter) net.IP {
	if addr, ok := w.RemoteAddr().(*net.TCPAddr); ok {
		return addr.IP
	} else if addr, ok := w.RemoteAddr().(*net.UDPAddr); ok {
		return addr.IP
	}
	panic("Needs to be either a TCP or UDP address")
}

func cleanup(db *sql.DB) {
	ctx := context.Background()
	_, span := tracer.Start(ctx, "cleanup")
	defer span.End()
	for {
		fmt.Println("Deleting old requests...")
		DeleteOldRequests(ctx, db)
		DeleteOldRecords(ctx, db)
		time.Sleep(time.Minute * 15)
	}
}

func lookupHost(ctx context.Context, ranges *Ranges, host net.IP) string {
	ctx, span := tracer.Start(ctx, "lookupHost")
	span.SetAttributes(attribute.String("host", host.String()))
	defer span.End()
	names, err := net.LookupAddr(host.String())
	if err == nil && len(names) > 0 {
		return names[0]
	}
	// otherwise search ASN database
	r, err := ranges.FindASN(host)
	if err != nil {
		return ""
	}
	return r.Name
}
