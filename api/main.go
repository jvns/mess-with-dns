package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
		fmt.Println("Error parsing record: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
        return
	}
	if !strings.HasSuffix(rr.Header().Name, ".messwithdns.com.") {
		fmt.Println("Invalid domain: ", rr.Header().Name)
		w.WriteHeader(http.StatusBadRequest)
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
	records := GetRecordsForName(db, domain + ".messwithdns.com.")
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
        http.ServeFile(w, r, "./frontend/" + r.URL.Path)
	}
}

func (handle *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)
	fmt.Println("Received request: ", r.Question[0].String())
	msg.Answer = GetRecords(handle.db, msg.Question[0].Name, msg.Question[0].Qtype)
	w.WriteMsg(&msg)
}

type UnknownRequest struct {
	Hdr dns.RR_Header
}

func main() {
	db := connect()
	handler := &handler{db: db}
	go func() {
		srv := &dns.Server{Handler: handler, Addr: ":5353", Net: "udp"}
		fmt.Println("Listening on :5353")
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
