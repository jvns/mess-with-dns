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

	"github.com/miekg/dns"
)

type handler struct {
	db *sql.DB
}

type RecordRequest struct {
	Domain string
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
	if !strings.HasSuffix(rr.Header().Name, ".messwithdns.com.") {
		fmt.Println("Invalid domain: ", rr.Header().Name)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	UpdateRecord(db, idInt, rr)
}

func getDomains(db *sql.DB, domain string, w http.ResponseWriter, r *http.Request) {
	// read body from json request
	records := GetRecordsForName(db, domain+".messwithdns.com.")
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
		getDomains(handle.db, p[1], w, r)
	// POST /record/new: add a new record
	case r.Method == "POST" && n == 2 && p[0] == "record" && p[1] == "new":
		createRecord(handle.db, w, r)
	// DELETE /record/<ID>:
	case r.Method == "DELETE" && n == 2 && p[0] == "record":
		deleteRecord(handle.db, p[1], w, r)
	// POST /record/<ID>: updates a record
	case r.Method == "POST" && n == 2 && p[0] == "record":
		updateRecord(handle.db, p[1], w, r)
	default:
		// serve static files
		http.ServeFile(w, r, "./frontend/"+r.URL.Path)
	}
}

func specialHandler(name string, qtype uint16) []dns.RR {
	if name == "messwithdns.com." && qtype == dns.TypeSOA {
		return []dns.RR{
			&dns.SOA{
				Hdr: dns.RR_Header{
					Name:   name,
					Rrtype: dns.TypeSOA,
					Class:  dns.ClassINET,
					Ttl:    0, /* RFC 1035 says soa records always should have a ttl of 0 */
				},
                Ns:      "ns1.messwithdns.com.",
                Mbox:    "julia.wizardzines.com.",
				Serial:  2,
				Refresh: 3600,
				Retry:   3600,
				Expire:  7300,
				Minttl:  5,
			},
		}
	}
    nameservers := []string{
        "213.188.214.254",
        "213.188.214.237",
    }
	if name == "ns1.messwithdns.com."  && qtype == dns.TypeA {
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
    if name == "ns2.messwithdns.com."  && qtype == dns.TypeA {
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

func (handle *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)
	msg.Authoritative = true
	fmt.Println("Received request: ", r.Question[0].String())
	specialRecords := specialHandler(r.Question[0].Name, r.Question[0].Qtype)
	if len(specialRecords) > 0 {
		msg.Answer = specialRecords
	} else {
		msg.Answer = GetRecords(handle.db, r.Question[0].Name, r.Question[0].Qtype)
	}
	err := w.WriteMsg(&msg)
	if err != nil {
		fmt.Println("Error writing response: ", err.Error())
		return
	}
	// print response
	if len(msg.Answer) > 0 {
		fmt.Println("Response: ", msg.Answer[0].String())
	} else {
		fmt.Println("Response: No records found")
	}
}

type UnknownRequest struct {
	Hdr dns.RR_Header
}

func main() {
	db := connect()
	handler := &handler{db: db}
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
	err := (&http.Server{Addr: ":8080", Handler: handler}).ListenAndServe()
	if err != nil {
		panic(err)
	}
}
