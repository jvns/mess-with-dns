package main

import (
	"encoding/json"
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

type handler struct{}

func (this *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
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
	jsonString := `{"Hdr":{"Name":"example.com.","Rrtype":15,"Class":1,"Ttl":3600,"Rdlength":0},"Preference":10,"Mx":"example.com."}`
	x, _ := ParseRecord(jsonString)
	fmt.Println(x.String())

	return

	rr, _ := dns.NewRR(fmt.Sprintf("example.com. IN MX 10 example.com"))
	// serialzie to json
	j, _ := json.Marshal(rr)
	// convert blah to bytes

	fmt.Println(string(j))
	fmt.Println(rr.String())

	srv := &dns.Server{Addr: ":" + strconv.Itoa(53), Net: "udp"}
	srv.Handler = &handler{}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to set udp listener %s\n", err.Error())
	}

}
