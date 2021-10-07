package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/miekg/dns"
)

var domainsToAddresses map[string]string = map[string]string{
	"google.com.":          "1.2.3.4",
	"jameshfisher.com.":    "104.198.14.52",
	"ns1.messwithdns.com.": "213.188.214.237",
	"ns2.messwithdns.com.": "213.188.214.254",
}

type handler struct {
	db *sql.DB
}

func (handle *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)
	fmt.Println("Received request: ", r.Question[0].String())
	switch r.Question[0].Qtype {
	case dns.TypeA:
		msg.Authoritative = true
		domain := msg.Question[0].Name
		address, ok := domainsToAddresses[domain]
		if ok {
			msg.Answer = append(msg.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
				A:   net.ParseIP(address),
			})
		}
	case dns.TypeNS:
		msg.Authoritative = true
		domain := msg.Question[0].Name
		address, ok := domainsToAddresses[domain]
		if ok {
			msg.Answer = append(msg.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
				A:   net.ParseIP(address),
			})
		}
	}
	w.WriteMsg(&msg)
}

type UnknownRequest struct {
	Hdr dns.RR_Header
}

func main() {
	db := connectDev()
	srv := &dns.Server{Addr: ":" + strconv.Itoa(53), Net: "udp"}
	srv.Handler = &handler{
		db: db,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to set udp listener %s\n", err.Error())
	}

}
