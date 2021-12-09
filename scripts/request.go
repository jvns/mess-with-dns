package main

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

func makeRequest() {
	// timeout after 1 second
	client := NewTimeoutClient(1*time.Second, 1*time.Second)
	resp, err := client.Get("http://mess-with-dns.fly.dev/bundle.js")
	// print time (not date)
	fmt.Print(time.Now().Format("15:04:05 "))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp.Status)
}

func main() {
	for {
		makeRequest()
		time.Sleep(1 * time.Second)
	}
}
func TimeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, nil
	}
}

func NewTimeoutClient(connectTimeout time.Duration, readWriteTimeout time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial: TimeoutDialer(500*time.Millisecond, 500*time.Millisecond),
		},
	}
}
