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
	domain := makeDomain(name)
	err := DeleteRequestsForDomain(db, domain)
	if err != nil {
		fmt.Println("Error deleting requests: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func getRequests(db *sql.DB, name string, w http.ResponseWriter, r *http.Request) {
	domain := makeDomain(name)
	requests, err := GetRequests(db, domain)
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

func createRecord(db *sql.DB, w http.ResponseWriter, r *http.Request) {
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
	if !validateDomainName(rr.Header().Name, w) {
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

func updateRecord(db *sql.DB, id string, w http.ResponseWriter, r *http.Request) {
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
	if !validateDomainName(rr.Header().Name, w) {
		return
	}
	UpdateRecord(db, idInt, rr)
}

func getDomains(db *sql.DB, name string, w http.ResponseWriter, r *http.Request) {
	domain := makeDomain(name)
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

func (handle *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request:", r.URL.Path)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

	p := strings.Split(r.URL.Path, "/")[1:]
	n := len(p)
	switch {
	// GET /domain/test : get everything from test.messwithdns.com.
	case r.Method == "GET" && n == 2 && p[0] == "domains":
		// set cache control header
		w.Header().Set("Cache-Control", "private, no-store")
		getDomains(handle.db, p[1], w, r)
	// GET /requests/test
	case r.Method == "GET" && n == 2 && p[0] == "requests":
		w.Header().Set("Cache-Control", "private, no-store")
		getRequests(handle.db, p[1], w, r)
	// DELETE /requests/test
	case r.Method == "DELETE" && n == 2 && p[0] == "requests":
		w.Header().Set("Cache-Control", "private, no-store")
		deleteRequests(handle.db, p[1], w, r)
	// GET /requeststream/test
	case r.Method == "GET" && n == 2 && p[0] == "requeststream":
		w.Header().Set("Cache-Control", "private, no-store")
		streamRequests(handle.db, p[1], w, r)
	// POST /record/new: add a new record
	case r.Method == "POST" && n == 2 && p[0] == "record" && p[1] == "new":
		w.Header().Set("Cache-Control", "private, no-store")
		createRecord(handle.db, w, r)
	// DELETE /record/<ID>:
	case r.Method == "DELETE" && n == 2 && p[0] == "record":
		w.Header().Set("Cache-Control", "private, no-store")
		deleteRecord(handle.db, p[1], w, r)
	// POST /record/<ID>: updates a record
	case r.Method == "POST" && n == 2 && p[0] == "record":
		w.Header().Set("Cache-Control", "private, no-store")
		updateRecord(handle.db, p[1], w, r)
	default:
		// serve static files
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
