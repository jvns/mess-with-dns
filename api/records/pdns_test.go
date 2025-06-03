package records_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/joeig/go-powerdns/v3"
	"github.com/jvns/mess-with-dns/records"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func setup() (records.RecordService, context.Context, string) {
	return records.Init("http://localhost:8082", "not-a-secret"), context.Background(), generateUsername()
}

func domain(username string) string {
	return username + ".messwithdns.com."
}

func generateUsername() string {
	// generate 16 digit random number
	username := "test_"
	for i := 0; i < 16; i++ {
		username += fmt.Sprintf("%d", rand.Intn(10))
	}
	return username
}

func sendDNSRequest(name string, qtype uint16) (*dns.Msg, error) {
	request := dns.Msg{
		Question: []dns.Question{
			{Name: name, Qtype: qtype, Qclass: dns.ClassINET},
		},
	}
	c := &dns.Client{Net: "udp"}
	response, _, err := c.Exchange(&request, "localhost:5555")
	return response, err
}

func TestCreateAndGetTxtRecords(t *testing.T) {
	rs, ctx, username := setup()
	record := map[string]string{"subdomain": "@", "type": "TXT", "ttl": "60", "value_Txt": "hello world"}
	err := rs.CreateRecord(ctx, username, record)
	if err != nil {
		t.Fatal(err)
	}

	records, err := rs.GetRecords(ctx, username)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("Expected 2 record, got %d", len(records))
	}
	assert.Equal(t, "@", records[0].Record.Subdomain)
	assert.Equal(t, "hello world", records[0].Record.Values["Txt"])
}

func TestCreateAndDeleteRecords(t *testing.T) {
	rs, ctx, username := setup()
	record := map[string]string{"subdomain": "@", "type": "A", "ttl": "60", "value_A": "1.2.3.4"}
	err := rs.CreateRecord(ctx, username, record)
	if err != nil {
		t.Fatal(err)
	}

	id := records.PdnsID{Name: domain(username), Type: powerdns.RRTypeA, Content: "1.2.3.4"}.String()
	err = rs.DeleteRecord(ctx, username, id)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateAndGetRecords(t *testing.T) {
	rs, ctx, username := setup()
	record := map[string]string{"subdomain": "@", "type": "A", "ttl": "60", "value_A": "1.2.3.4"}
	err := rs.CreateRecord(ctx, username, record)
	if err != nil {
		t.Fatal(err)
	}

	records, err := rs.GetRecords(ctx, username)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("Expected 2 record, got %d", len(records))
	}
	assert.Equal(t, "@", records[0].Record.Subdomain)
}

func TestCreateAndUpdateRecord(t *testing.T) {
	rs, ctx, username := setup()
	record := map[string]string{"subdomain": "@", "type": "A", "ttl": "60", "value_A": "1.2.3.4"}
	err := rs.CreateRecord(ctx, username, record)
	if err != nil {
		t.Fatal(err)
	}

	id := records.PdnsID{Name: fmt.Sprintf("%s.messwithdns.com.", username), Type: powerdns.RRTypeA, Content: "1.2.3.4"}.String()
	err = rs.UpdateRecord(ctx, username, id, map[string]string{"subdomain": "@", "type": "A", "ttl": "60", "value_A": "2.3.4.5"})
	if err != nil {
		t.Fatal(err)
	}

	records, err := rs.GetRecords(ctx, username)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 2, len(records))
	assert.Equal(t, records[len(records)-1].Record.Values["A"], "2.3.4.5")
}

func TestCreateAndUpdateRecordName(t *testing.T) {
	rs, ctx, username := setup()
	record := map[string]string{"subdomain": "@", "type": "A", "ttl": "60", "value_A": "1.2.3.4"}
	err := rs.CreateRecord(ctx, username, record)
	if err != nil {
		t.Fatal(err)
	}

	id := records.PdnsID{Name: fmt.Sprintf("%s.messwithdns.com.", username), Type: powerdns.RRTypeA, Content: "1.2.3.4"}.String()
	err = rs.UpdateRecord(ctx, username, id, map[string]string{"subdomain": "test", "type": "A", "ttl": "60", "value_A": "2.3.4.5"})
	if err != nil {
		t.Fatal(err)
	}

	records, err := rs.GetRecords(ctx, username)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 2, len(records))
	assert.Equal(t, "test", records[0].Record.Subdomain)
}

// test updating a record with another record with the same name, different case
func TestCreateAndUpdateRecordCase(t *testing.T) {
	rs, ctx, username := setup()
	record := map[string]string{"subdomain": "banana", "type": "A", "ttl": "60", "value_A": "1.2.3.4"}
	err := rs.CreateRecord(ctx, username, record)
	if err != nil {
		t.Fatal(err)
	}
	id := records.PdnsID{Name: fmt.Sprintf("banana.%s.messwithdns.com.", username), Type: powerdns.RRTypeA, Content: "1.2.3.4"}.String()

	err = rs.UpdateRecord(ctx, username, id, map[string]string{"subdomain": "BANANA", "type": "A", "ttl": "60", "value_A": "1.2.3.4"})
	if err != nil {
		t.Fatal(err)
	}

	records, err := rs.GetRecords(ctx, username)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 2, len(records))
	assert.Equal(t, "banana", records[0].Record.Subdomain)
}

func TestCreateManyRecords(t *testing.T) {
	rs, ctx, username := setup()
	records := []map[string]string{
		{"subdomain": "@", "type": "A", "ttl": "60", "value_A": "1.2.3.4"},
		{"subdomain": "@", "type": "AAAA", "ttl": "60", "value_AAAA": "2001:db8::1"},
		{"subdomain": "@", "type": "CAA", "ttl": "60", "value_Flag": "0", "value_Tag": "issue", "value_Value": "letsencrypt.org"},
		{"subdomain": "test", "type": "CNAME", "ttl": "60", "value_Target": "example.com"},
		{"subdomain": "@", "type": "MX", "ttl": "60", "value_Preference": "10", "value_Mx": "example.com"},
		{"subdomain": "@", "type": "PTR", "ttl": "60", "value_Ptr": "example.com"},
		{"subdomain": "@", "type": "TXT", "ttl": "60", "value_Txt": "hello world"},
		{"subdomain": "@", "type": "SRV", "ttl": "60", "value_Priority": "10", "value_Weight": "10", "value_Port": "8080", "value_Target": "orange-ip.fly.dev"},
		{"subdomain": "@", "type": "NS", "ttl": "60", "value_Ns": "ns1.example.com"},
		// SVCB, with/without params
		{"subdomain": "@", "type": "SVCB", "ttl": "60", "value_Priority": "10", "value_Target": "example.com"},
		{"subdomain": "@", "type": "SVCB", "ttl": "60", "value_Priority": "10", "value_Target": "example.com", "value_Params": "alpn=h3"},
		// HTTPS, with/without params
		{"subdomain": "@", "type": "HTTPS", "ttl": "60", "value_Priority": "10", "value_Target": "example.com"},
		{"subdomain": "@", "type": "HTTPS", "ttl": "60", "value_Priority": "10", "value_Target": "example.com", "value_Params": "alpn=h3"},
		// spaces should get trimmed
		{"subdomain": "orange ", "type": "A", "ttl": "60", "value_A": "1.2.3.4 "},
		// underscore is okay
		{"subdomain": "_test", "type": "A", "ttl": "60", "value_A": "1.2.3.4"},
		// SRV record with priority 0 and weight 0 is ok
		{"subdomain": "@", "type": "SRV", "ttl": "60", "value_Priority": "0", "value_Weight": "0", "value_Port": "8080", "value_Target": "orange-ip.fly.dev"},
		// dashes are okay
		{"subdomain": "test-dash.test-dash", "type": "A", "ttl": "60", "value_A": "1.2.3.4"},
	}
	for _, record := range records {
		// generate a new username each time so that CNAME doesn't conflict
		err := rs.CreateRecord(ctx, username, record)
		if err != nil {
			t.Fatal(err)
		}
	}
}

type ErrorTest struct {
	Record map[string]string
	Error  string
}

func TestTranslateErrors(t *testing.T) {
	rs, ctx, username := setup()
	tests := []ErrorTest{
		{map[string]string{"subdomain": "new site", "type": "A", "ttl": "60", "value_A": "1.2.3.4"}, "Error: name \"new site.%s.messwithdns.com.\" contains a space"},
	}
	for _, test := range tests {
		err := rs.CreateRecord(ctx, username, test.Record)
		formattedError := fmt.Sprintf(test.Error, username)
		assert.Equal(t, formattedError, err.Error())
	}
}

func TestCreateRecordFixMX(t *testing.T) {
	rs, ctx, username := setup()
	record := map[string]string{"subdomain": "@", "type": "MX", "ttl": "60", "value_Preference": "10", "value_Mx": "example.com"}
	err := rs.CreateRecord(ctx, username, record)
	if err != nil {
		t.Fatal(err)
	}

	records, err := rs.GetRecords(ctx, username)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 2, len(records))
	assert.Equal(t, "example.com.", records[0].Record.Values["Mx"])
}

func TestPunycode(t *testing.T) {
	rs, ctx, username := setup()
	record := map[string]string{"subdomain": "❤", "type": "A", "ttl": "60", "value_A": "1.2.3.4"}
	err := rs.CreateRecord(ctx, username, record)
	if err != nil {
		t.Fatal(err)
	}

	response, err2 := sendDNSRequest(fmt.Sprintf("xn--qei.%s.messwithdns.com.", username), dns.TypeA)
	assert.Nil(t, err2)
	assert.Equal(t, len(response.Answer), 1)
}

func TestDNSQueryNoRecord(t *testing.T) {
	response, err2 := sendDNSRequest("example.com.", dns.TypeA)
	if err2 != nil {
		t.Fatal(err2)
	}

	assert.Equal(t, 0, len(response.Answer))
}

func TestDNSQuery(t *testing.T) {
	rs, ctx, username := setup()
	record := map[string]string{"subdomain": "@", "type": "A", "ttl": "60", "value_A": "1.2.3.4"}
	err := rs.CreateRecord(ctx, username, record)
	if err != nil {
		t.Fatal(err)
	}

	response, err2 := sendDNSRequest(domain(username), dns.TypeA)
	if err2 != nil {
		t.Fatal(err2)
	}

	assert.Equal(t, response.Answer[0].Header().Name, domain(username))
	assert.Equal(t, response.Answer[0].Header().Rrtype, dns.TypeA)
	assert.Equal(t, response.Answer[0].Header().Ttl, uint32(60))
	assert.Equal(t, response.Answer[0].(*dns.A).A.String(), "1.2.3.4")
}

func TestParseSerial(t *testing.T) {
	var serial uint32 = 2021091008
	serialInt, err := records.ParseSerial(serial)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, time.Date(2021, 9, 10, 0, 0, 0, 0, time.UTC), serialInt)
}
