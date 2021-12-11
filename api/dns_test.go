package main

import (
	"database/sql"
	"net"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

// some integration-style tests
var connString = "postgres://postgres:mysecretpassword@localhost:5432/postgres?sslmode=disable"

func connectTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func makeA(name string, ip string) *dns.A {
	return &dns.A{
		Hdr: dns.RR_Header{
			Name:   name,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    0,
		},
		A: net.ParseIP(ip),
	}
}

func makeCNAME(name string, target string) *dns.CNAME {
	return &dns.CNAME{
		Hdr: dns.RR_Header{
			Name:   name,
			Rrtype: dns.TypeCNAME,
			Class:  dns.ClassINET,
			Ttl:    0,
		},
		Target: target,
	}
}

func makeQuestion(name string, qtype uint16) *dns.Msg {
	return &dns.Msg{
		Question: []dns.Question{
			{
				Name:   name,
				Qtype:  qtype,
				Qclass: dns.ClassINET,
			},
		},
	}
}

func TestARecord(t *testing.T) {
	db := connectTestDB(t)
	name := randString(10) + ".messwithdns.com."
	InsertRecord(db, makeA(name, "1.2.3.4"))
	response := dnsResponse(db, makeQuestion(name, dns.TypeA))
	// check that we got NOERROR and 1 answer
	assert.Equal(t, dns.RcodeSuccess, response.Rcode)
	assert.Equal(t, 1, len(response.Answer))
}

func TestCNAMERecord(t *testing.T) {
	db := connectTestDB(t)
	name := randString(10) + ".messwithdns.com."
	InsertRecord(db, makeCNAME(name, "example.com."))
	response := dnsResponse(db, makeQuestion(name, dns.TypeA))
	// check that we got NOERROR and 1 answer
	assert.Equal(t, dns.RcodeSuccess, response.Rcode)
	assert.Equal(t, 1, len(response.Answer))
}

func TestHTTPSRecord(t *testing.T) {
	db := connectTestDB(t)
	name := randString(10) + ".messwithdns.com."
	InsertRecord(db, makeA(name, "1.2.3.4"))
	response := dnsResponse(db, makeQuestion(name, dns.TypeHTTPS))
	// check that we got NOERROR and 1 answer
	assert.Equal(t, dns.RcodeSuccess, response.Rcode)
	assert.Equal(t, 1, len(response.Answer))
}

func TestNoError(t *testing.T) {
	db := connectTestDB(t)
	name := randString(10) + ".messwithdns.com."
	InsertRecord(db, makeA(name, "1.2.3.4"))
	response := dnsResponse(db, makeQuestion(name, dns.TypeAAAA))
	// check that we got NOERROR and 0 answers
	assert.Equal(t, dns.RcodeSuccess, response.Rcode)
	assert.Equal(t, 0, len(response.Answer))
}

func TestNXDOMAIN(t *testing.T) {
	db := connectTestDB(t)
	name := randString(10) + ".messwithdns.com."
	response := dnsResponse(db, makeQuestion(name, dns.TypeA))
	// check that we got NXDOMAIN
	assert.Equal(t, dns.RcodeNameError, response.Rcode)
}
