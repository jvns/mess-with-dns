package main

import (
	"database/sql"
	"fmt"
	"net"
	"strings"

	"github.com/miekg/dns"
)

func lookupRecords(db *sql.DB, name string, qtype uint16) ([]dns.RR, int, error) {
	records := specialRecords(name, qtype)
	if len(records) > 0 {
		return records, len(records), nil
	}
	return GetRecords(db, name, qtype)
}

func dnsResponse(db *sql.DB, request *dns.Msg) *dns.Msg {
	if !strings.HasSuffix(strings.ToLower(request.Question[0].Name), "messwithdns.com.") {
		return refusedResponse(request)
	}
	records, totalRecords, err := lookupRecords(db, request.Question[0].Name, request.Question[0].Qtype)
	if err != nil {
		msg := errorResponse(request)
		fmt.Println("Error getting records:", err)
		return msg
	}
	if totalRecords == 0 {
		return nxDomainResponse(request)
	}
	return successResponse(request, records)
}

func emptyMessage(request *dns.Msg) *dns.Msg {
	msg := dns.Msg{Compress: true}
	msg.SetReply(request)
	msg.Authoritative = true
	msg.Ns = []dns.RR{
		getSOA(soaSerial),
	}
	return &msg
}

func errorResponse(request *dns.Msg) *dns.Msg {
	msg := emptyMessage(request)
	msg.SetRcode(request, dns.RcodeServerFailure)
	return msg
}

func nxDomainResponse(request *dns.Msg) *dns.Msg {
	msg := emptyMessage(request)
	msg.SetRcode(request, dns.RcodeNameError)
	return msg
}

func refusedResponse(request *dns.Msg) *dns.Msg {
	msg := dns.Msg{Compress: true}
	msg.SetReply(request)
	msg.Authoritative = true

	msg.SetRcode(request, dns.RcodeRefused)
	return &msg
}

func successResponse(request *dns.Msg, records []dns.RR) *dns.Msg {
	msg := emptyMessage(request)
	msg.Answer = records
	return msg
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
	"www.messwithdns.com": &dns.A{
		Hdr: dns.RR_Header{
			Name:   "messwithdns.com.",
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    60,
		},
		A: net.ParseIP("213.188.214.254"),
	},
	"messwithdns.com.": &dns.A{
		Hdr: dns.RR_Header{
			Name:   "messwithdns.com.",
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    60,
		},
		A: net.ParseIP("213.188.214.254"),
	},
	"_psl.messwithdns.com.": &dns.TXT{
		Hdr: dns.RR_Header{
			Name:   "_psl.messwithdns.com.",
			Rrtype: dns.TypeTXT,
			Class:  dns.ClassINET,
			Ttl:    60,
		},
		Txt: []string{"https://github.com/publicsuffix/list/pull/1490"},
	},
}

func specialRecords(name string, qtype uint16) []dns.RR {
	if record, ok := records[name]; ok {
		if record.Header().Rrtype == qtype {
			return []dns.RR{record}
		}
	}
	// special case for SOA
	if qtype == dns.TypeSOA && name == "messwithdns.com." {
		return []dns.RR{getSOA(soaSerial)}
	}
	if qtype == dns.TypeNS && name == "messwithdns.com." {
		return []dns.RR{
			&dns.NS{
				Hdr: dns.RR_Header{
					Name:   "messwithdns.com.",
					Rrtype: dns.TypeNS,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				Ns: "mess-with-dns1.wizardzines.com.",
			},
			&dns.NS{
				Hdr: dns.RR_Header{
					Name:   "messwithdns.com.",
					Rrtype: dns.TypeNS,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				Ns: "mess-with-dns2.wizardzines.com.",
			},
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
		Ns:      "mess-with-dns1.wizardzines.com.",
		Mbox:    "julia.wizardzines.com.",
		Serial:  serial,
		Refresh: 3600,
		Retry:   3600,
		Expire:  7300,
		Minttl:  3600, // MINIMUM is a lower bound on the TTL field for all RRs in a zone
	}
	return &soa
}
