package parsing

import (
	//"encoding/json"
	//"github.com/miekg/dns"
	//"net"
	"testing"

	"github.com/joeig/go-powerdns/v3"
	"github.com/stretchr/testify/assert"
)

func fatalIfErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseMX(t *testing.T) {
	record := map[string]string{"subdomain": "@", "type": "MX", "ttl": "3600", "value_Preference": "10", "value_Mx": "mail.example.com."}
	x, err := ParseRecordRequest(record, "test")
	fatalIfErr(t, err)
	assert.Equal(t, *x.Records[0].Content, "10 mail.example.com.")
	assert.Equal(t, *x.Type, powerdns.RRTypeMX)
	assert.Equal(t, *x.Name, "test.messwithdns.com.")
	assert.Equal(t, *x.TTL, uint32(3600))
}

func TestParseTxt(t *testing.T) {
	record := map[string]string{"subdomain": "@", "type": "TXT", "ttl": "60", "value_Txt": "hello world"}
	x, err := ParseRecordRequest(record, "test")
	fatalIfErr(t, err)
	assert.Equal(t, *x.Records[0].Content, "\"hello world\"")
}

type TestCase struct {
	Record  map[string]string
	Content string
}

func TestParseAll(t *testing.T) {
	testCases := []TestCase{
		{
			Record:  map[string]string{"subdomain": "@", "type": "A", "ttl": "3600", "value_A": "1.2.3.4"},
			Content: "1.2.3.4",
		},
		{
			Record:  map[string]string{"subdomain": "@", "type": "AAAA", "ttl": "3600", "value_AAAA": "2001:db8::1"},
			Content: "2001:db8::1",
		},
		{
			Record:  map[string]string{"subdomain": "@", "type": "CNAME", "ttl": "3600", "value_Target": "example.com."},
			Content: "example.com.",
		},
		{
			Record:  map[string]string{"subdomain": "@", "type": "MX", "ttl": "3600", "value_Preference": "10", "value_Mx": "mail.example.com."},
			Content: "10 mail.example.com.",
		},
		{
			// same but with no trailing dot
			Record:  map[string]string{"subdomain": "@", "type": "MX", "ttl": "3600", "value_Preference": "10", "value_Mx": "mail.example.com"},
			Content: "10 mail.example.com.",
		},
		{
			Record:  map[string]string{"subdomain": "@", "type": "NS", "ttl": "3600", "value_Ns": "ns1.example.com."},
			Content: "ns1.example.com.",
		},
		{
			Record:  map[string]string{"subdomain": "@", "type": "PTR", "ttl": "3600", "value_Ptr": "www.example.com."},
			Content: "www.example.com.",
		},
		{
			Record:  map[string]string{"subdomain": "@", "type": "SRV", "ttl": "3600", "value_Priority": "10", "value_Weight": "10", "value_Port": "8080", "value_Target": "orange-ip.fly.dev."},
			Content: "10 10 8080 orange-ip.fly.dev.",
		},
		{
			Record:  map[string]string{"subdomain": "@", "type": "TXT", "ttl": "3600", "value_Txt": "hello world"},
			Content: "\"hello world\"",
		},
		{
			Record:  map[string]string{"subdomain": "@", "type": "CAA", "ttl": "3600", "value_Flag": "1", "value_Tag": "issue", "value_Value": "ca.example.com"},
			Content: "1 issue \"ca.example.com\"",
		},
		// SOA
		{
			Record:  map[string]string{"subdomain": "@", "type": "SOA", "ttl": "3600", "value_Mname": "ns1.example.com.", "value_Rname": "hostmaster@example.com.", "value_Serial": "2021010101", "value_Refresh": "3600", "value_Retry": "600", "value_Expire": "604800", "value_Minimum": "3600"},
			Content: "ns1.example.com. hostmaster.example.com. 2021010101 3600 600 604800 3600",
		},
	}

	for _, testCase := range testCases {
		x, err := ParseRecordRequest(testCase.Record, "test")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, *x.Records[0].Content, testCase.Content)
		assert.Equal(t, *x.Name, "test.messwithdns.com.")
		assert.Equal(t, *x.TTL, uint32(3600))
		assert.Equal(t, *x.Type, powerdns.RRType(testCase.Record["type"]))
	}
}

func TestParseError(t *testing.T) {
	// Some invalid records, maybe add more later
	testCases := []map[string]string{
		{"subdomain": "@", "type": "A", "ttl": "3600", "value_A": "1.2.3.5555"},
		{"subdomain": "@", "type": "A", "ttl": "3600", "value_A": "banana"},
	}

	for _, testCase := range testCases {
		_, err := ParseRecordRequest(testCase, "test")
		if err == nil {
			t.Fatal("expected error for", testCase)
		}
	}
}
