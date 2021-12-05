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

type RecordRequest struct {
	Domain string
}

var soaSerial uint32

func getSOA(serial uint32) *dns.SOA {
	var soa = dns.SOA{
		Hdr: dns.RR_Header{
			Name:   "messwithdns.com.",
			Rrtype: dns.TypeSOA,
			Class:  dns.ClassINET,
			Ttl:    300, /* RFC 1035 says soa records always should have a ttl of 0 but cloudflare doesn't seem to do that*/
		},
		Ns:      "ns1.messwithdns.com.",
		Mbox:    "julia.wizardzines.com.",
		Serial:  serial,
		Refresh: 3600,
		Retry:   3600,
		Expire:  7300,
		Minttl:  3600, // MINIMUM is a lower bound on the TTL field for all RRs in a zone
	}
	return &soa
}

func deleteRequests(db *sql.DB, name string, w http.ResponseWriter, r *http.Request) {
	domain := name + ".messwithdns.com."
	err := DeleteRequestsForDomain(db, domain)
	if err != nil {
		fmt.Println("Error deleting requests: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func getRequests(db *sql.DB, name string, w http.ResponseWriter, r *http.Request) {
	domain := name + ".messwithdns.com."
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
	domain := name + ".messwithdns.com."
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
		errMsg := fmt.Sprintf("Error parsing record: %s", err.Error())
		fmt.Println(errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errMsg))
		return
	}
	if !strings.HasSuffix(rr.Header().Name, ".messwithdns.com.") {
		errMsg := fmt.Sprintf("Invalid domain: %s", rr.Header().Name)
		fmt.Println(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		return
	}
	if !validateSubdomain(rr.Header().Name, w) {
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
	// make sure update subdomain is valid
	if !validateSubdomain(rr.Header().Name, w) {
		return
	}
	if err != nil {
		fmt.Println("Error parsing record: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
	if !strings.HasSuffix(rr.Header().Name, ".messwithdns.com.") {
		fmt.Println("Invalid domain: ", rr.Header().Name)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	UpdateRecord(db, idInt, rr)
}

func getDomains(db *sql.DB, name string, w http.ResponseWriter, r *http.Request) {
	domain := name + ".messwithdns.com."
	// read body from json request
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

func specialHandler(db *sql.DB, name string, qtype uint16) []dns.RR {
	if name == "messwithdns.com." && qtype == dns.TypeSOA {
		return []dns.RR{
			getSOA(soaSerial),
		}
	}
	nameservers := []string{
		"213.188.214.254",
		"213.188.214.237",
	}
	if name == "fly-test." && qtype == dns.TypeA {
		return []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    3600,
				},
				A: net.ParseIP("1.2.3.4"),
			},
		}
	}
	if name == "orange.messwithdns.com." && qtype == dns.TypeA {
		return []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    3600,
				},
				A: net.ParseIP("213.188.218.160"),
			},
		}
	}
	if name == "purple.messwithdns.com." && qtype == dns.TypeA {
		return []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    3600,
				},
				A: net.ParseIP("213.188.209.192"),
			},
		}
	}
	if name == "www.messwithdns.com." && qtype == dns.TypeA {
		return []dns.RR{
			&dns.CNAME{
				Hdr: dns.RR_Header{
					Name:   name,
					Rrtype: dns.TypeCNAME,
					Class:  dns.ClassINET,
					Ttl:    3600,
				},
				Target: "mess-with-dns.fly.dev.",
			},
		}
	}
	if name == "ns1.messwithdns.com." && qtype == dns.TypeA {
		return []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    3600,
				},
				A: net.ParseIP(nameservers[0]),
			},
		}
	}
	if name == "ns2.messwithdns.com." && qtype == dns.TypeA {
		return []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    3600,
				},
				A: net.ParseIP(nameservers[1]),
			},
		}
	}

	return nil
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

func (handle *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	// get time
	start := time.Now()
	msg := dns.Msg{}
	msg.SetReply(r)
	msg.Authoritative = true
	fmt.Println("Received request: ", r.Question[0].String())
	specialRecords := specialHandler(handle.db, r.Question[0].Name, r.Question[0].Qtype)
	if len(specialRecords) > 0 {
		msg.Answer = specialRecords
	} else {
		answer, err := GetRecords(handle.db, r.Question[0].Name, r.Question[0].Qtype)
		if err != nil {
			// internal server error
			msg := dns.Msg{}
			msg.SetRcode(r, dns.RcodeServerFailure)
			w.WriteMsg(&msg)
			fmt.Println("Error getting records:", err)
			return
		}
		msg.Answer = answer
	}
	// add SOA record
	msg.Ns = []dns.RR{
		getSOA(soaSerial),
	}
	err := w.WriteMsg(&msg)
	if err != nil {
		fmt.Println("Error writing response: ", err.Error())
		return
	}
	// print response
	// get time since start
	elapsed := time.Since(start)
	if len(msg.Answer) > 0 {
		fmt.Println("Response: ", msg.Answer[0].String(), elapsed)
	} else {
		// print elapsed time
		fmt.Println("Response: (no records found)", elapsed)

	}
	remote_addr := w.RemoteAddr().(*net.UDPAddr).IP
	LogRequest(handle.db, r, &msg, remote_addr, lookupHost(handle.ipRanges, remote_addr))
}
