package main

import (
	"database/sql"
	"net"

	"github.com/miekg/dns"
)

func lookupRecords(db *sql.DB, name string, qtype uint16) ([]dns.RR, error) {
	records := specialRecords(name, qtype)
	if len(records) > 0 {
		return records, nil
	}
	return GetRecords(db, name, qtype)
}

func errorResponse(request *dns.Msg) *dns.Msg {
	msg := dns.Msg{}
	msg.SetReply(request)
	msg.Authoritative = true
	msg.SetRcode(nil, dns.RcodeServerFailure)
	msg.Ns = []dns.RR{
		getSOA(soaSerial),
	}
	return &msg
}

func successResponse(request *dns.Msg, records []dns.RR) *dns.Msg {
	msg := dns.Msg{}
	msg.SetReply(request)
	msg.Authoritative = true
	msg.Answer = records
	msg.Ns = []dns.RR{
		getSOA(soaSerial),
	}
	return &msg
}

var records = map[string]dns.RR{
	"fly-test.": &dns.A{
		Hdr: dns.RR_Header{
			Name:   "fly-test.",
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    60,
		},
		A: net.ParseIP("1.2.3.4"),
	},
	"orange.messwithdns.com.": &dns.A{
		Hdr: dns.RR_Header{
			Name:   "orange.messwithdns.com.",
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    3600,
		},
		A: net.ParseIP("213.188.218.160"),
	},
	"purple.messwithdns.com.": &dns.A{
		Hdr: dns.RR_Header{
			Name:   "purple.messwithdns.com.",
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    3600,
		},
		A: net.ParseIP("213.188.209.192"),
	},
	"ns1.messwithdns.com.": &dns.A{
		Hdr: dns.RR_Header{
			Name:   "ns1.messwithdns.com.",
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    3600,
		},
		A: net.ParseIP("213.188.214.254"),
	},
	"ns2.messwithdns.com.": &dns.A{
		Hdr: dns.RR_Header{
			Name:   "ns2.messwithdns.com.",
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    3600,
		},
		A: net.ParseIP("213.188.214.237"),
	},
	"www.messwithdns.com": &dns.CNAME{
		Hdr: dns.RR_Header{
			Name:   "www.messwithdns.com.",
			Rrtype: dns.TypeCNAME,
			Class:  dns.ClassINET,
			Ttl:    3600,
		},
		Target: "mess-with-dns.fly.dev.",
	},
	"messwithdns.com.": getSOA(soaSerial),
}

func specialRecords(name string, qtype uint16) []dns.RR {
	if record, ok := records[name]; ok {
		if record.Header().Rrtype == qtype {
			return []dns.RR{record}
		}
	}
	return nil
}

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
