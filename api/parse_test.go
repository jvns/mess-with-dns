package main

import (
	"testing"
    "net"

	"github.com/stretchr/testify/assert"
)

func TestParseMX(t *testing.T) {
	jsonString := `{"Hdr":{"Name":"example.com.","Rrtype":15,"Class":1,"Ttl":3600,"Rdlength":0},"Preference":10,"Mx":"mail.example.com."}`
	x, _ := ParseRecord([]byte(jsonString))
	assert.Equal(t, x.String(), "example.com.	3600	IN	MX	10 mail.example.com.")
}

func TestInvalidFqdn(t *testing.T) {
	jsonString := `{"Hdr":{"Name":"example.com.","Rrtype":15,"Class":1,"Ttl":3600,"Rdlength":0},"Preference":10,"Mx":"mail.example.com"}`
	_, err := ParseRecord([]byte(jsonString))
	assert.Equal(t, err.Error(), "Invalid RR: dns: domain must be fully qualified, &dns.MX{Hdr:dns.RR_Header{Name:\"example.com.\", Rrtype:0xf, Class:0x1, Ttl:0xe10, Rdlength:0x0}, Preference:0xa, Mx:\"mail.example.com\"}")
}

func TestParseASN(t *testing.T) {
    ranges, _ := ReadASNs("../ip2asn-v4.tsv")
    r, _ := FindASN(ranges, net.ParseIP("172.217.13.174"))
    assert.Equal(t, r.Num, 15169)
    assert.Equal(t, r.Name, "Google LLC")

    _, err := FindASN(ranges, net.ParseIP("255.255.255.255"))
    assert.Equal(t, err.Error(), "not found")

    _, err = FindASN(ranges, net.ParseIP("0.0.0.0"))
    assert.Equal(t, err.Error(), "not found")
}
