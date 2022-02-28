package main

// domain -> stream id -> channel

import (
	"math/rand"
)

var streams = map[string]map[string]chan []byte{}

type Stream struct {
	id        string
	subdomain string
}

func CreateStream(subdomain string) Stream {
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

func WriteToStreams(domain string, msg []byte) {
	if _, ok := streams[domain]; ok {
		for _, stream := range streams[domain] {
			stream <- msg
		}
	}
}
