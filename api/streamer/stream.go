package streamer

import (
	"encoding/json"
	"github.com/miekg/dns"
	"math/rand"
	"net"
	"strings"
	"time"
)

var streams = map[string]map[string]chan []byte{}

type Stream struct {
	id        string
	subdomain string
}

func (l *Logger) CreateStream(subdomain string) Stream {
	if _, ok := streams[subdomain]; !ok {
		streams[subdomain] = make(map[string]chan []byte)
	}
	id := randString(10)
	streams[subdomain][id] = make(chan []byte)
	return Stream{id: id, subdomain: subdomain}
}

func randString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (s *Stream) Delete() {
	if _, ok := streams[s.subdomain]; ok {
		close(streams[s.subdomain][s.id])
		delete(streams[s.subdomain], s.id)
	}
}

func (s *Stream) Get() chan []byte {
	if _, ok := streams[s.subdomain]; ok {
		return streams[s.subdomain][s.id]
	}
	return nil
}

func writeToStreams(domain string, response *dns.Msg, src_host string, src_ip net.IP) error {
	streamLog := responseToStreamLog(time.Now().Unix(), response, src_host, src_ip.String())
	msg, err := json.Marshal(streamLog)
	if err != nil {
		return err
	}

	domain = strings.ToLower(domain)
	if _, ok := streams[domain]; ok {
		for _, stream := range streams[domain] {
			stream <- msg
		}
	}
	return nil
}
