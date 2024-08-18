package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/jvns/mess-with-dns/records"
	"github.com/jvns/mess-with-dns/streamer"
	"github.com/jvns/mess-with-dns/users"
	"github.com/miekg/dns"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"io"
	"net/http"
	"time"
)

func getRecords(ctx context.Context, username string, rs records.RecordService, w http.ResponseWriter, r *http.Request) {
	records, err := rs.GetRecords(ctx, username)
	if err != nil {
		returnError(w, r, err, err.Code)
		return
	}
	jsonOutput, err2 := json.Marshal(records)
	if err != nil {
		returnError(w, r, err2, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonOutput)
}

func deleteRecord(ctx context.Context, username string, id string, rs records.RecordService, w http.ResponseWriter, r *http.Request) {
	err := rs.DeleteRecord(ctx, username, id)
	if err != nil {
		returnError(w, r, err, err.Code)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func updateRecord(ctx context.Context, username string, id string, rs records.RecordService, w http.ResponseWriter, r *http.Request) {
	record := map[string]string{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		returnError(w, r, fmt.Errorf("error reading body: %s", err.Error()), http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &record); err != nil {
		returnError(w, r, fmt.Errorf("error decoding json: %s, body: %s", err.Error(), string(body)), http.StatusBadRequest)
		return
	}
	err2 := rs.UpdateRecord(ctx, username, id, record)
	if err2 != nil {
		returnError(w, r, err2, err2.Code)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func createRecord(ctx context.Context, username string, rs records.RecordService, w http.ResponseWriter, r *http.Request) {
	record := map[string]string{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		returnError(w, r, fmt.Errorf("error reading body: %s", err.Error()), http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &record); err != nil {
		returnError(w, r, fmt.Errorf("error decoding json: %s, body: %s", err.Error(), string(body)), http.StatusBadRequest)
		return
	}
	err2 := rs.CreateRecord(ctx, username, record)
	if err2 != nil {
		returnError(w, r, err2, err2.Code)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func deleteRequests(ctx context.Context, logger *streamer.Logger, name string, w http.ResponseWriter, r *http.Request) {
	err := logger.DeleteRequestsForDomain(ctx, name)
	if err != nil {
		err := fmt.Errorf("error deleting requests: %s", err.Error())
		returnError(w, r, err, http.StatusInternalServerError)
		return
	}
}

func getRequests(ctx context.Context, logger *streamer.Logger, username string, w http.ResponseWriter, r *http.Request) {
	requests, err := logger.GetRequests(ctx, username)
	if err != nil {
		err := fmt.Errorf("error getting requests: %s", err.Error())
		returnError(w, r, err, http.StatusInternalServerError)
		return
	}
	jsonOutput, err := json.Marshal(requests)
	if err != nil {
		err := fmt.Errorf("error marshalling json: %s", err.Error())
		returnError(w, r, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonOutput)
}

func streamRequests(ctx context.Context, logger *streamer.Logger, subdomain string, w http.ResponseWriter, r *http.Request) {
	// create websocket connection
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		returnError(w, r, err, http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	if err != nil {
		err := fmt.Errorf("error creating websocket connection: %s", err.Error())
		returnError(w, r, err, http.StatusInternalServerError)
		return
	}
	stream := logger.CreateStream(subdomain)
	defer stream.Delete()
	c := stream.Get()
	// I don't really understand this ping/pong stuff but it's what the gorilla docs say to do
	ticker := time.NewTicker(15 * time.Second)
	pongWait := time.Second * 60
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		select {
		case <-ticker.C:
			err := conn.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				// I think this just means the client disconnected
				return
			}
		case msg := <-c:
			span := trace.SpanFromContext(ctx)
			span.SetAttributes(attribute.String("stream.request", string(msg)))
			err := conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				span.RecordError(err)
				fmt.Println("Error writing message:", err)
				span.End()
				return
			}
			span.End()
		}
	}
}

func loginRandom(u *users.UserService, rs records.RecordService, w http.ResponseWriter, r *http.Request) {
	subdomain, err := u.CreateAvailableSubdomain()

	if err != nil {
		returnError(w, r, err, http.StatusInternalServerError)
		return
	}
	u.SetCookie(w, r, subdomain)
	rs.CreateZone(context.Background(), subdomain)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	// make dns request to check if we're up
	m := new(dns.Msg)
	m.SetQuestion("glass99.messwithdns.com.", dns.TypeA)
	m.RecursionDesired = true
	c := new(dns.Client)
	c.DialTimeout = time.Second * 1
	//c.Net = "tcp"
	response, _, err := c.Exchange(m, "127.0.0.1:53")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error making DNS request: " + err.Error()))
		return
	}
	// Make sure that response code is NXDOMAIN
	if response.Rcode != dns.RcodeNameError {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Expected NXDOMAIN, got " + dns.RcodeToString[response.Rcode]))
	}

	/*
		// get requests for glass99 and make sure it's a 200 ok
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://127.0.0.1:8080/requests", nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error creating request"))
			return
		}
		req.Header.Set("User-Agent", "healthcheck")
		req.Header.Set("Cookie", "session=MTY1MDEyMjU3MnxTcnB5M3ZvYmFKRXBhWXV0Y3kwWWNTTk5mU05Nb3hvRG5yajNkM2Fod2dNRVJ3MEJUX0RwTng2anduVGpOYVdTVENTdFY3aXNPWEJxVUNORXJlSGp8vaZ4BQTLfPwl6xy5VIvMQsqB2qiTjgss2RYWJUqCCTM=; username=glass99")
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != 200 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error requesting /requests"))
			return
		}
	*/
}
