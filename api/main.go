package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/miekg/dns"
)

func main() {
	db, err := connect()
	if err != nil {
		panic(fmt.Sprintf("Error connecting to database: %s", err.Error()))
	}
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
		panic(err)
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

func deleteRequests(db *sql.DB, name string, w http.ResponseWriter, r *http.Request) {
	err := DeleteRequestsForDomain(db, name)
	if err != nil {
		fmt.Println("Error deleting requests: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func getRequests(db *sql.DB, username string, w http.ResponseWriter, r *http.Request) {
	requests, err := GetRequests(db, username)
	if err != nil {
		fmt.Println("Error getting requests: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	jsonOutput, err := json.Marshal(requests)
	if err != nil {
		fmt.Println("Error marshalling json: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonOutput)
}

func streamRequests(db *sql.DB, name string, w http.ResponseWriter, r *http.Request) {
	// create websocket connection
	conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if err != nil {
		fmt.Println("Error upgrading websocket: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
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
		fmt.Println("Error reading body: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	rr, err := ParseRecord(body)
	if err != nil {
		var errMsg string
		if strings.Contains(err.Error(), "Invalid RR") {
			errMsg = "Oops, invalid domain name"
		} else {
			errMsg = fmt.Sprintf("Error parsing record: %s", err.Error())
		}
		fmt.Println(errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errMsg))
		return
	}
	if !validateDomainName(rr.Header().Name, username, w) {
		return
	}
	InsertRecord(db, rr)
}

func deleteRecord(db *sql.DB, id string, w http.ResponseWriter, r *http.Request) {
	// parse int from id
	idInt, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println("Error parsing id: ", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	DeleteRecord(db, idInt)
}

func updateRecord(db *sql.DB, username string, id string, w http.ResponseWriter, r *http.Request) {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println("Error parsing id: ", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Error reading body: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
	rr, err := ParseRecord(body)
	if err != nil {
		fmt.Println("Error parsing record: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
	if !validateDomainName(rr.Header().Name, username, w) {
		return
	}
	UpdateRecord(db, idInt, rr)
}

func getDomains(db *sql.DB, username string, w http.ResponseWriter, r *http.Request) {
	domain := makeDomain(username)
	records, err := GetRecordsForName(db, domain)
	if err != nil {
		fmt.Println("Error getting records: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	jsonOutput, err := json.Marshal(records)
	if err != nil {
		fmt.Println("Error marshalling json: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonOutput)
}

func requireLogin(username string, w http.ResponseWriter) bool {
	w.Header().Set("Cache-Control", "no-store")
	if username == "" {
		w.WriteHeader(http.StatusUnauthorized)
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
		githubOauth(w)
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
