package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	//"github.com/jvns/mess-with-dns/streamer"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func fatalIfErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func createTestHandler(t *testing.T) *handler {
	base64Hash := "/JLayjTcQf0wl/YifN7WqyP6U1+y/qnxxNzhbQ1Falk="
	base64Block := "SaJ+upj49i3BzLP46bUh5g860DgB+V5z4zuTlevI9ug="
	config := &Config{
		workdir:           "..",
		requestDBFilename: ":memory:",
		userDBFilename:    ":memory:",
		hashKey:           base64Hash,
		blockKey:          base64Block,
		powerdnsAddress:   "http://localhost:8082",
		dnstapAddress:     "localhost:7111",
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	handler, err := createHandler(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	return handler
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

func createTestServer(t *testing.T) *httptest.Server {
	handler := createTestHandler(t)
	return httptest.NewServer(createRoutes(handler))
}

/* test harness time */

func createTestHarness(t *testing.T) *TestHarness {
	ts := createTestServer(t)
	client, ws, username := login(t, ts)
	return &TestHarness{t: t, ts: ts, client: client, ws: ws, username: username}
}

type TestHarness struct {
	t        *testing.T
	ts       *httptest.Server
	client   *http.Client
	ws       *websocket.Conn
	username string
}

func (h *TestHarness) CreateRecord(jsonString string) {
	t := h.t
	resp, err := h.client.Post(h.ts.URL+"/records", "application/json", strings.NewReader(jsonString))
	fatalIfErr(t, err)
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Error %d: %s", resp.StatusCode, body)
	}
}

func (h *TestHarness) UpdateRecord(id string, jsonString string) {
	t := h.t
	resp, err := h.client.Post(h.ts.URL+"/records/"+id, "application/json", strings.NewReader(jsonString))
	fatalIfErr(t, err)
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Error %d: %s", resp.StatusCode, body)
	}
}

func (h *TestHarness) GetRecords() []Record2 {
	t := h.t
	resp, err := h.client.Get(h.ts.URL + "/records")
	fatalIfErr(t, err)
	assert.Equal(t, resp.StatusCode, 200)
	body, _ := io.ReadAll(resp.Body)
	var records []Record2
	err = json.Unmarshal(body, &records)
	fatalIfErr(t, err)
	return records
}

func (h *TestHarness) Domain() string {
	return fmt.Sprintf("%s.messwithdns.com.", h.username)
}

func TestLogin(t *testing.T) {
	ts := createTestServer(t)
	defer ts.Close()
	client := ts.Client()
	// don't check redirects
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resp, err := client.Get(ts.URL + "/login")
	fatalIfErr(t, err)
	if resp.StatusCode != 307 {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Error %d: %s", resp.StatusCode, body)
		return
	}
	cookies := resp.Cookies()
	assert.Equal(t, len(cookies), 2)
	assert.Equal(t, cookies[0].Name, "session")
	assert.Equal(t, cookies[1].Name, "username")

}

func login(t *testing.T, ts *httptest.Server) (*http.Client, *websocket.Conn, string) {
	client := ts.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resp, err := client.Get(ts.URL + "/login")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 307 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Error %d: %s", resp.StatusCode, body)
	}
	cookies := resp.Cookies()
	username := cookies[1].Value
	if cookies[1].Name != "username" {
		t.Fatalf("Expected username cookie, got %s", cookies[1].Name)
	}
	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
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
		t.Fatal(err)
	}
	ws.SetReadDeadline(time.Now().Add(1 * time.Second))

	return client, ws, username
}

type Record2 struct {
	ID     string            `json:"id"`
	Record map[string]string `json:"record"`
}

/*
func TestWebsocket(t *testing.T) {
	h := createTestHarness(t)

	jsonString := `{"subdomain":"@","type":"A","ttl":"60","value_A":"1.2.3.4"}`
	h.CreateRecord(jsonString)

	sendDNSRequest(h.Domain(), dns.TypeA)

	_, message, err := h.ws.ReadMessage()
	fatalIfErr(t, fmt.Errorf("Error reading websocket message: %s", err))
	var log streamer.StreamLog
	err = json.Unmarshal(message, &log)
	fatalIfErr(t, err)
	assert.Equal(t, log.Response.Code, "NOERROR")
	// check content
	if len(log.Response.Records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(log.Response.Records))
	}
	assert.Equal(t, log.Response.Records[0].Content, "1.2.3.4")
}

/*

func TestGetRequests(t *testing.T) {
	ts, dnst := createTestServer()
	client, _, username, err := login(ts)
	fatalIfErr(t, err)
	jsonString := `{"subdomain":"@","type":"A","ttl":60,"values":[{"name":"A", "value":"1.2.3.4"}]}`
	resp, err := client.Post(ts.URL+"/records", "application/json", strings.NewReader(jsonString))
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
	resp, err = client.Get(ts.URL + "/requests")
	fatalIfErr(t, err)
	body, err := io.ReadAll(resp.Body)
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
	resp, err := client.Post(ts.URL+"/records", "application/json", strings.NewReader(jsonString))
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
*/
