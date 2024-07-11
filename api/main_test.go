package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/miekg/dns"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/jvns/mess-with-dns/dnstester"
	"github.com/jvns/mess-with-dns/parsing"
	"github.com/stretchr/testify/assert"
)

func fatalIfErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func createTestHandler() *handler {
	db, err := connect()
	if err != nil {
		panic(fmt.Sprintf("Error connecting to database: %s", err.Error()))
	}
	err = createTables(db)
	if err != nil {
		panic(fmt.Sprintf("Error creating tables: %s", err.Error()))
	}
	go cleanup(db)
	ctx := context.Background()
	soaSerial, err = GetSerial(ctx, db)
	if err != nil {
		panic(fmt.Sprintf("Error getting SOA serial: %s", err.Error()))
	}
	ranges, err := parsing.ReadRanges()
	if err != nil {
		panic(fmt.Sprintf("Error reading ranges: %s", err.Error()))
	}
	return &handler{db: db, ipRanges: &ranges}
}

func createTestServer() (*httptest.Server, *dnstester.DNSTester) {
	handler := createTestHandler()
	return httptest.NewServer(handler), dnstester.NewDNSTester(handler)
}

func TestLogin(t *testing.T) {
	ts, _ := createTestServer()
	defer ts.Close()
	client := ts.Client()
	// don't check redirects
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resp, err := client.Get(ts.URL + "/login")
	fatalIfErr(t, err)
	if resp.StatusCode != 307 {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Errorf("Error %d: %s", resp.StatusCode, body)
		return
	}
	cookies := resp.Cookies()
	assert.Equal(t, len(cookies), 2)
	assert.Equal(t, cookies[0].Name, "session")
	assert.Equal(t, cookies[1].Name, "username")

}

func login(ts *httptest.Server) (*http.Client, *websocket.Conn, string, error) {
	client := ts.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resp, err := client.Get(ts.URL + "/login")
	if err != nil {
		return nil, nil, "", err
	}
	if resp.StatusCode != 307 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, nil, "", fmt.Errorf("Error %d: %s", resp.StatusCode, body)
	}
	cookies := resp.Cookies()
	username := cookies[1].Value
	if cookies[1].Name != "username" {
		return nil, nil, "", fmt.Errorf("Expected username cookie, got %s", cookies[1].Name)
	}
	u, err := url.Parse(ts.URL)
	if err != nil {
		return nil, nil, "", err
	}
	jar, _ := cookiejar.New(nil)
	jar.SetCookies(u, cookies)
	client.Jar = jar

	// open websocket
	wsURL := strings.Replace(ts.URL, "http", "ws", 1) + "/requeststream"
	dialer := websocket.Dialer{
		Jar: jar,
	}
	ws, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		return nil, nil, "", err
	}
	ws.SetReadDeadline(time.Now().Add(1 * time.Second))

	return client, ws, username, nil
}

func TestCreateRecord(t *testing.T) {
	ts, _ := createTestServer()
	client, _, _, err := login(ts)
	fatalIfErr(t, err)
	jsonString := `{"subdomain":"@","type":"A","ttl":60,"values":[{"name":"A", "value":"1.2.3.4"}]}`
	resp, err := client.Post(ts.URL+"/records/", "application/json", strings.NewReader(jsonString))
	fatalIfErr(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestCreateRecordFixMX(t *testing.T) {
	// it should add a '.' to the end
	ts, dnst := createTestServer()
	client, _, username, err := login(ts)
	fatalIfErr(t, err)
	jsonString := `{"subdomain":"@","type":"MX","ttl":60,"values":[{"name": "Preference", "value": "10"}, {"name": "Mx", "value": "example.com"}]}`
	resp, err := client.Post(ts.URL+"/records/", "application/json", strings.NewReader(jsonString))
	fatalIfErr(t, err)
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("Error %d: %s", resp.StatusCode, body)
	}

	response := dnst.Request(&dns.Msg{
		Question: []dns.Question{
			{Name: fmt.Sprintf("%s.messwithdns.com.", username), Qtype: dns.TypeMX, Qclass: dns.ClassINET},
		},
	})
	if len(response.Answer) != 1 {
		t.Fatalf("Expected 1 answer, got %d", len(response.Answer))
	}
	assert.Equal(t, response.Answer[0].(*dns.MX).Mx, "example.com.")
}

func TestPunycode(t *testing.T) {
	ts, dnst := createTestServer()
	client, _, username, err := login(ts)
	fatalIfErr(t, err)

	// create a record with a punycode domain
	jsonString := `{"subdomain":"❤","type":"A","ttl":60,"values":[{"name":"A", "value":"1.2.3.4"}]}`
	resp, err := client.Post(ts.URL+"/records/", "application/json", strings.NewReader(jsonString))
	fatalIfErr(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	response := dnst.Request(&dns.Msg{
		Question: []dns.Question{
			{Name: fmt.Sprintf("xn--qei.%s.messwithdns.com.", username), Qtype: dns.TypeA, Qclass: dns.ClassINET},
		},
	})
	assert.Equal(t, len(response.Answer), 1)
}

func TestCreateAndGetRecords(t *testing.T) {
	ts, _ := createTestServer()
	client, _, _, err := login(ts)
	fatalIfErr(t, err)
	jsonString := `{"subdomain":"@","type":"A","ttl":60,"values":[{"name":"A", "value":"1.2.3.4"}]}`
	resp, err := client.Post(ts.URL+"/records/", "application/json", strings.NewReader(jsonString))
	fatalIfErr(t, err)

	resp, err = client.Get(ts.URL + "/records/")
	fatalIfErr(t, err)
	assert.Equal(t, resp.StatusCode, 200)
	body, _ := ioutil.ReadAll(resp.Body)
	var records []Record
	err = json.Unmarshal(body, &records)
	fatalIfErr(t, err)
	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}
	assert.Equal(t, "@", records[0].Record.Subdomain)
}

func TestCreateAndGetTxtRecords(t *testing.T) {
	ts, _ := createTestServer()
	client, _, _, err := login(ts)
	fatalIfErr(t, err)
	jsonString := `{"subdomain":"@","type":"TXT","ttl":60,"values":[{"name":"Txt", "value":"hello world"}]}`
	resp, err := client.Post(ts.URL+"/records/", "application/json", strings.NewReader(jsonString))
	fatalIfErr(t, err)

	resp, err = client.Get(ts.URL + "/records/")
	fatalIfErr(t, err)
	assert.Equal(t, resp.StatusCode, 200)
	body, _ := ioutil.ReadAll(resp.Body)
	var records []Record
	err = json.Unmarshal(body, &records)
	fatalIfErr(t, err)
	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}
	assert.Equal(t, "@", records[0].Record.Subdomain)
	assert.Equal(t, "hello world", records[0].Record.Values[0].Value)
}

func TestDNSQuery(t *testing.T) {
	ts, dnst := createTestServer()
	client, _, username, err := login(ts)
	fatalIfErr(t, err)
	jsonString := `{"subdomain":"@","type":"A","ttl":60,"values":[{"name":"A", "value":"1.2.3.4"}]}`
	resp, err := client.Post(ts.URL+"/records/", "application/json", strings.NewReader(jsonString))
	fatalIfErr(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	response := dnst.Request(&dns.Msg{
		Question: []dns.Question{
			{Name: fmt.Sprintf("%s.messwithdns.com.", username), Qtype: dns.TypeA, Qclass: dns.ClassINET},
		},
	})

	if len(response.Answer) != 1 {
		t.Fatalf("Expected 1 answer, got %d", len(response.Answer))
	}

	assert.Equal(t, response.Answer[0].Header().Name, fmt.Sprintf("%s.messwithdns.com.", username))
	assert.Equal(t, response.Answer[0].Header().Rrtype, dns.TypeA)
	assert.Equal(t, response.Answer[0].Header().Ttl, uint32(60))
	assert.Equal(t, response.Answer[0].(*dns.A).A.String(), "1.2.3.4")
}

func TestDNSQueryNoRecord(t *testing.T) {
	_, dnst := createTestServer()
	response := dnst.Request(&dns.Msg{
		Question: []dns.Question{
			{Name: "example.com.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
		},
	})

	assert.Equal(t, len(response.Answer), 0)
}

func TestWebsocket(t *testing.T) {
	ts, dnst := createTestServer()
	client, ws, username, err := login(ts)
	defer ws.Close()
	fatalIfErr(t, err)
	jsonString := `{"subdomain":"@","type":"A","ttl":60,"values":[{"name":"A", "value":"1.2.3.4"}]}`
	resp, err := client.Post(ts.URL+"/records/", "application/json", strings.NewReader(jsonString))
	fatalIfErr(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	dnst.Request(&dns.Msg{
		Question: []dns.Question{
			{Name: fmt.Sprintf("%s.messwithdns.com.", username), Qtype: dns.TypeA, Qclass: dns.ClassINET},
		},
	})
	_, message, err := ws.ReadMessage()
	fatalIfErr(t, err)
	var log StreamLog
	err = json.Unmarshal(message, &log)
	fatalIfErr(t, err)
	assert.Equal(t, log.Response.Code, "NOERROR")
	// check content
	if len(log.Response.Records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(log.Response.Records))
	}
	assert.Equal(t, log.Response.Records[0].Content, "1.2.3.4")

}

func TestGetRequests(t *testing.T) {
	ts, dnst := createTestServer()
	client, _, username, err := login(ts)
	fatalIfErr(t, err)
	jsonString := `{"subdomain":"@","type":"A","ttl":60,"values":[{"name":"A", "value":"1.2.3.4"}]}`
	resp, err := client.Post(ts.URL+"/records/", "application/json", strings.NewReader(jsonString))
	fatalIfErr(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	msg := &dns.Msg{
		Question: []dns.Question{
			{Name: fmt.Sprintf("%s.messwithdns.com.", username), Qtype: dns.TypeA, Qclass: dns.ClassINET},
		},
	}
	dnst.Request(msg)
	dnst.Request(msg)
	fatalIfErr(t, err)

	// get /requests
	resp, err = client.Get(ts.URL + "/requests/")
	fatalIfErr(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	fatalIfErr(t, err)
	if resp.StatusCode != 200 {
		t.Fatal(string(body))
	}
	logs := []StreamLog{}
	if err = json.Unmarshal(body, &logs); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, logs[0].Response.Code, "NOERROR")
	assert.Equal(t, len(logs[0].Response.Records), 1)

}

func TestWebsocketMixedCaseRequest(t *testing.T) {
	ts, dnst := createTestServer()
	client, ws, username, err := login(ts)
	defer ws.Close()
	fatalIfErr(t, err)
	jsonString := `{"subdomain":"@","type":"A","ttl":60,"values":[{"name":"A", "value":"1.2.3.4"}]}`
	resp, err := client.Post(ts.URL+"/records/", "application/json", strings.NewReader(jsonString))
	fatalIfErr(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	dnst.Request(&dns.Msg{
		Question: []dns.Question{
			{Name: fmt.Sprintf("%s.mESSWIthDnS.com.", username), Qtype: dns.TypeA, Qclass: dns.ClassINET},
		},
	})
	_, message, err := ws.ReadMessage()
	if err != nil {
		assert.Fail(t, err.Error())
		return
	}
	assert.True(t, strings.Contains(string(message), username))
}
