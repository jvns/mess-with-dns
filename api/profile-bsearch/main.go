package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net/netip"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/jvns/mess-with-dns/streamer/ip2asn"
)

func memusage() {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MiB\n", m.Alloc/1024/1024)
	// write mem.prof
	f, err := os.Create("mem.prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.WriteHeapProfile(f)
	f.Close()
}

func main() {
	ranges, err := ip2asn.ReadRanges("..")
	if err != nil {
		log.Fatal(err)
	}
	memusage()

	// try it 1000000 times
	ips := []netip.Addr{}
	for i := 0; i < 100000; i++ {
		// get 2 random 64-bit integers
		u1 := rand.Uint64()
		u2 := rand.Uint64()
		// create a random IPv6 address
		ip := netip.AddrFrom16([16]byte{uint8(u1), uint8(u1 >> 8), uint8(u1 >> 16), uint8(u1 >> 24), uint8(u1 >> 32), uint8(u1 >> 40), uint8(u1 >> 48), uint8(u1 >> 56), uint8(u2), uint8(u2 >> 8), uint8(u2 >> 16), uint8(u2 >> 24), uint8(u2 >> 32), uint8(u2 >> 40), uint8(u2 >> 48), uint8(u2 >> 56)})
		ips = append(ips, ip)
	}
	now := time.Now()
	success := 0
	for _, ip := range ips {
		_, err := ranges.FindASN(ip)
		if err == nil {
			success++
		}
	}
	fmt.Println(success)
	elapsed := time.Since(now)
	fmt.Println("number per second", float64(success)/elapsed.Seconds())
}
