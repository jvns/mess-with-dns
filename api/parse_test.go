package main

import "testing"

func TestParseMX(t *testing.T) {
	jsonString := `{"Hdr":{"Name":"example.com.","Rrtype":15,"Class":1,"Ttl":3600,"Rdlength":0},"Preference":10,"Mx":"mail.example.com."}`
	x, _ := ParseRecord(jsonString)
	if x.String() != "example.com.	3600	IN	MX	10 mail.example.com." {
		t.Errorf("mx record didn't match %s", x.String())
	}
}
