package main

import (
	"testing"

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
	assert.Equal(t, err.Error(), "Invalid RR: dns: domain must be fully qualified")
}
