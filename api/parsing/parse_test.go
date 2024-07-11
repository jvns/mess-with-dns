package parsing

import (
	"encoding/json"
	"github.com/miekg/dns"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMX(t *testing.T) {
	jsonString := `{"subdomain":"@","type":"MX","ttl":3600,"values":[{"name": "Preference", "value": "10"}, {"name": "Mx", "value": "mail.example.com."}]}`
	x, err := ParseJSRecord([]byte(jsonString), "test")
	fatalIfErr(t, err)
	assert.Equal(t, x.String(), "test.messwithdns.com.	3600	IN	MX	10 mail.example.com.")
}

func TestParseTxt(t *testing.T) {
	jsonString := `{"subdomain":"@","type":"TXT","ttl":3600,"values":[{"name":"Txt","value":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}]}`
	x, err := ParseJSRecord([]byte(jsonString), "test")
	fatalIfErr(t, err)
	assert.Equal(t, x.String(), "test.messwithdns.com.	3600	IN	TXT	\"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\" \"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\"")
}

func TestParseMXSubdomain(t *testing.T) {
	jsonString := `{"subdomain":"blob","type":"MX","ttl":3600,"values":[{"name": "Preference", "value": "10"}, {"name": "Mx", "value": "mail.example.com."}]}`
	x, err := ParseJSRecord([]byte(jsonString), "test")
	fatalIfErr(t, err)
	assert.Equal(t, x.String(), "blob.test.messwithdns.com.	3600	IN	MX	10 mail.example.com.")
}

func TestInvalidFqdn(t *testing.T) {
	jsonString := `{"Hdr":{"Name":"example.com.","Rrtype":15,"Class":1,"Ttl":3600,"Rdlength":0},"Preference":10,"Mx":"mail.example.com"}`
	_, err := ParseRecord([]byte(jsonString))
	assert.Equal(t, err.Error(), "Invalid RR: dns: domain must be fully qualified, &dns.MX{Hdr:dns.RR_Header{Name:\"example.com.\", Rrtype:0xf, Class:0x1, Ttl:0xe10, Rdlength:0x0}, Preference:0xa, Mx:\"mail.example.com\"}")
}

func TestParseInvalidFqdnJS(t *testing.T) {
	jsonString := `{"subdomain":"blob","type":"MX","ttl":3600,"values":[{"name": "Preference", "value": "10"}, {"name": "Mx", "value": "mail.example.com."}]}`
	_, err := ParseJSRecord([]byte(jsonString), "test")
	assert.Nil(t, err)
}

func TestParseASN(t *testing.T) {
	ranges, _ := ReadASNs("../../ip2asn-v4.tsv")
	r, _ := FindASN(ranges, net.ParseIP("172.217.13.174"))
	assert.Equal(t, r.Num, 15169)
	assert.Equal(t, r.Name, "GOOGLE")

	_, err := FindASN(ranges, net.ParseIP("255.255.255.255"))
	assert.Equal(t, err.Error(), "not found")

	_, err = FindASN(ranges, net.ParseIP("0.0.0.0"))
	assert.Equal(t, err.Error(), "not found")
}

// now for the RR to js record part

func TestBasicRR(t *testing.T) {
	record := "test.messwithdns.com. 50 IN A 1.2.3.4"
	rr, err := dns.NewRR(record)
	fatalIfErr(t, err)
	js, err := RRToJSRecord(rr)
	fatalIfErr(t, err)
	assert.Equal(t, js.Subdomain, "@")
	assert.Equal(t, js.Typ, "A")
	assert.Equal(t, js.TTL, uint32(50))
	assert.Equal(t, js.Values[0].Value, "1.2.3.4")
}

func TestTxtRR(t *testing.T) {
	record := "test.messwithdns.com. 50 IN TXT \"hello\""
	rr, err := dns.NewRR(record)
	fatalIfErr(t, err)
	js, err := RRToJSRecord(rr)
	fatalIfErr(t, err)
	assert.Equal(t, js.Values[0].Value, "hello")
}

func fatalIfErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

var testRecords []string = []string{
	"test.messwithdns.com. 50 IN A 1.2.3.4",
	"test.messwithdns.com. 50 IN AAAA 2001:db8::1",
	"test.messwithdns.com. 50 IN CNAME example.com.",
	"test.messwithdns.com. 50 IN MX 10 mail.example.com.",
	"test.messwithdns.com. 50 IN NS ns1.example.com.",
	"test.messwithdns.com. 50 IN PTR www.example.com.",
	"test.messwithdns.com. 50 IN SRV 10 10 8080 orange-ip.fly.dev.",
	"test.messwithdns.com. 50 IN TXT \"hello world\"",
	"test.messwithdns.com. 50 IN CAA 1 issue \"ca.example.com\"",
}

func TestRoundTrip(t *testing.T) {
	for _, record := range testRecords {
		rr, err := dns.NewRR(record)
		fatalIfErr(t, err)
		js, err := RRToJSRecord(rr)
		fatalIfErr(t, err)
		json, err := json.Marshal(js)
		fatalIfErr(t, err)
		rr2, err := ParseJSRecord(json, "test")
		fatalIfErr(t, err)
		assert.Equal(t, rr.String(), rr2.String())
	}
}

func TestSchemaOrder(t *testing.T) {
	schemas, err := GenerateSchemas()
	fatalIfErr(t, err)
	for _, record := range testRecords {
		rr, err := dns.NewRR(record)
		fatalIfErr(t, err)
		js, err := RRToJSRecord(rr)
		// test that the order in the values matches the order in the schema

		schema, ok := schemas[js.Typ]
		if !ok {
			t.Fatalf("no schema for %s", js.Typ)
		}
		for i, value := range js.Values {
			assert.Equal(t, schema[i].Name, value.Name)
		}
	}

}
