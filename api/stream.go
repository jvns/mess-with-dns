package main

// domain -> stream id -> channel

import (
	"fmt"
	"math/rand"
)

var streams = map[string]map[string]chan []byte{}

type Stream struct {
	id     string
	domain string
}

func CreateStream(domain string) Stream {
	if _, ok := streams[domain]; !ok {
		streams[domain] = make(map[string]chan []byte)
	}
	id := randString(10)
	streams[domain][id] = make(chan []byte)
	return Stream{id: id, domain: domain}
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
	if _, ok := streams[s.domain]; ok {
		close(streams[s.domain][s.id])
		delete(streams[s.domain], s.id)
	}
}

func (s *Stream) Get() chan []byte {
	if _, ok := streams[s.domain]; ok {
		return streams[s.domain][s.id]
	}
	return nil
}

func WriteToStreams(domain string, msg []byte) {
	if _, ok := streams[domain]; ok {
		for _, stream := range streams[domain] {
			fmt.Println("writing to stream")
			stream <- msg
		}
	}
}
