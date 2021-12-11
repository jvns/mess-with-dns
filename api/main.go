package main

import (
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

	"github.com/getsentry/sentry-go"
	"github.com/gorilla/websocket"
	"github.com/miekg/dns"
)

func main() {
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
	soaSerial, err = GetSerial(db)
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
	fmt.Println("Listening on :8080")
	err = (&http.Server{Addr: ":8080", Handler: handler}).ListenAndServe()
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

func returnError(w http.ResponseWriter, err error, status int) {
	fmt.Printf("Error [%d]: %s\n", status, err.Error())
	sentry.CaptureException(err)
	http.Error(w, err.Error(), status)
}

func deleteRequests(db *sql.DB, name string, w http.ResponseWriter, r *http.Request) {
	err := DeleteRequestsForDomain(db, name)
	if err != nil {
		err := fmt.Errorf("Error deleting requests: %s", err.Error())
		returnError(w, err, http.StatusInternalServerError)
		return
	}
}

func getRequests(db *sql.DB, username string, w http.ResponseWriter, r *http.Request) {
	requests, err := GetRequests(db, username)
	if err != nil {
		err := fmt.Errorf("Error getting requests: %s", err.Error())
		returnError(w, err, http.StatusInternalServerError)
		return
	}
	jsonOutput, err := json.Marshal(requests)
	if err != nil {
		err := fmt.Errorf("Error marshalling json: %s", err.Error())
		returnError(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonOutput)
}

func streamRequests(db *sql.DB, name string, w http.ResponseWriter, r *http.Request) {
	// create websocket connection
	conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if err != nil {
		err := fmt.Errorf("Error creating websocket connection: %s", err.Error())
		returnError(w, err, http.StatusInternalServerError)
		return
	}
	domain := makeDomain(name)
	stream := CreateStream(domain)
	defer stream.Delete()
	c := stream.Get()
	for {
		select {
		case msg := <-c:
			conn.WriteMessage(websocket.TextMessage, msg)
		}
	}
}

func createRecord(db *sql.DB, username string, w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err := fmt.Errorf("Error reading body: %s", err.Error())
		returnError(w, err, http.StatusInternalServerError)
		return
	}
	rr, err := ParseRecord(body)
	if err != nil {
		if strings.Contains(err.Error(), "Invalid RR") {
			returnError(w, fmt.Errorf("Oops, invalid domain name"), http.StatusBadRequest)
		} else {
			returnError(w, fmt.Errorf("Error parsing record: %s", err.Error()), http.StatusInternalServerError)
		}
		return
	}
	if err = validateDomainName(rr.Header().Name, username); err != nil {
		returnError(w, err, http.StatusBadRequest)
		return
	}
	InsertRecord(db, rr)
}

func deleteRecord(db *sql.DB, id string, w http.ResponseWriter, r *http.Request) {
	// parse int from id
	idInt, err := strconv.Atoi(id)
	if err != nil {
		err := fmt.Errorf("Error parsing id: %s", err.Error())
		returnError(w, err, http.StatusBadRequest)
		return
	}
	DeleteRecord(db, idInt)
}

func updateRecord(db *sql.DB, username string, id string, w http.ResponseWriter, r *http.Request) {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		returnError(w, fmt.Errorf("Error parsing id: %s", err.Error()), http.StatusBadRequest)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		returnError(w, fmt.Errorf("Error reading body: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	rr, err := ParseRecord(body)
	if err != nil {
		returnError(w, fmt.Errorf("Error parsing record: %s", err.Error()), http.StatusBadRequest)
		return
	}
	if err = validateDomainName(rr.Header().Name, username); err != nil {
		returnError(w, err, http.StatusBadRequest)
		return
	}
	UpdateRecord(db, idInt, rr)
}

func getDomains(db *sql.DB, username string, w http.ResponseWriter, r *http.Request) {
	domain := makeDomain(username)
	records, err := GetRecordsForName(db, domain)
	if err != nil {
		returnError(w, fmt.Errorf("Error getting records: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	jsonOutput, err := json.Marshal(records)
	if err != nil {
		returnError(w, fmt.Errorf("Error marshalling json: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonOutput)
}

func requireLogin(username string, w http.ResponseWriter) bool {
	w.Header().Set("Cache-Control", "no-store")
	if username == "" {
		returnError(w, fmt.Errorf("You must be logged in to access this page"), http.StatusUnauthorized)
		return false
	}
	return true
}

func (handle *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request:", r.URL.Path)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")
	username, _ := ReadSessionUsername(r)

	p := strings.Split(r.URL.Path, "/")[1:]
	n := len(p)
	switch {
	// GET /domain: get everything from USERNAME.messwithdns.com.
	case r.Method == "GET" && p[0] == "domains":
		if !requireLogin(username, w) {
			return
		}
		getDomains(handle.db, username, w, r)
	// GET /requests
	case r.Method == "GET" && p[0] == "requests":
		if !requireLogin(username, w) {
			return
		}
		getRequests(handle.db, username, w, r)
	// DELETE /requests
	case r.Method == "DELETE" && p[0] == "requests":
		if !requireLogin(username, w) {
			return
		}
		deleteRequests(handle.db, username, w, r)
	// GET /requeststream
	case r.Method == "GET" && p[0] == "requeststream":
		if !requireLogin(username, w) {
			return
		}
		streamRequests(handle.db, username, w, r)
	// POST /record/new: add a new record
	case r.Method == "POST" && n == 2 && p[0] == "record" && p[1] == "new":
		if !requireLogin(username, w) {
			return
		}
		createRecord(handle.db, username, w, r)
	// DELETE /record/<ID>:
	case r.Method == "DELETE" && n == 2 && p[0] == "record":
		if !requireLogin(username, w) {
			return
		}
		// TODO: don't let people delete other people's records
		deleteRecord(handle.db, p[1], w, r)
	// POST /record/<ID>: updates a record
	case r.Method == "POST" && n == 2 && p[0] == "record":
		if !requireLogin(username, w) {
			return
		}
		updateRecord(handle.db, username, p[1], w, r)
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
	start := time.Now()
	fmt.Println("Received request: ", r.Question[0].String())

	records, err := lookupRecords(handle.db, r.Question[0].Name, r.Question[0].Qtype)
	if err != nil {
		msg := errorResponse(r)
		fmt.Println("Error getting records:", err)
		w.WriteMsg(msg)
		return
	}
	msg := successResponse(r, records)
	elapsed := time.Since(start)
	if len(msg.Answer) > 0 {
		fmt.Println("Response: ", msg.Answer[0].String(), elapsed)
	} else {
		// print elapsed time
		fmt.Println("Response: (no records found)", elapsed)

	}
	w.WriteMsg(msg)
	remote_addr := w.RemoteAddr().(*net.UDPAddr).IP
	LogRequest(handle.db, r, msg, remote_addr, lookupHost(handle.ipRanges, remote_addr))
}

func cleanup(db *sql.DB) {
	for {
		fmt.Println("Deleting old requests...")
		DeleteOldRequests(db)
		DeleteOldRecords(db)
		time.Sleep(time.Minute * 15)
	}
}

func lookupHost(ranges *Ranges, host net.IP) string {
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
