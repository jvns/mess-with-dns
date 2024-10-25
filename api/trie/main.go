package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"

	"github.com/seancfoley/ipaddress-go/ipaddr"
)

func main() {
	// Create a new ASN trie
	lookup := NewASNTrie()

	filename := "../ip2asn-v4.tsv"
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(f)

	// Parse and add each record
	for scanner.Scan() {
		err := lookup.AddRecord(scanner.Text())
		if err != nil {
			fmt.Printf("Error adding record: %v\n", err)
			continue
		}
	}

	// Example lookups
	testIPs := []string{
		"1.0.0.100",
		"1.0.2.50",
		"8.8.8.8",
		"9.9.9.9", // Should not find a match
	}

	for _, ip := range testIPs {
		_, _ = lookup.LookupIP(ip)
		//if err != nil {
		//	fmt.Printf("Lookup for %s: %v\n", ip, err)
		//	continue
		//}
		//fmt.Printf("IP: %s\nASN: %d\nCountry: %s\nNetwork: %s\n\n",
		//	ip, info.ASN, info.Country, info.Network)
	}
	// print programs' memory usage

	fmt.Println("Memory Usage")

	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	// current heap size
	fmt.Printf("\tHeapAlloc = %v MiB", bToMb(m.HeapAlloc))
	// run a GC

	// write pprof dat to file
	f, err = os.Create("mem.pprof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.WriteHeapProfile(f)

}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// ASNInfo holds the metadata for an IP range
type ASNInfo struct {
	ASN     int
	Country string
	Network string
}

// ASNTrie wraps the library's trie for ASN lookups
type ASNTrie struct {
	trie *ipaddr.AssociativeTrie[*ipaddr.IPAddress, *ASNInfo]
}

// NewASNTrie creates a new ASN lookup trie
func NewASNTrie() *ASNTrie {
	return &ASNTrie{
		trie: ipaddr.NewAssociativeTrie[*ipaddr.IPAddress, *ASNInfo](),
	}
}

// AddRecord parses a line of ASN data and adds it to the trie
func (t *ASNTrie) AddRecord(line string) error {
	// Split the line into fields
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return fmt.Errorf("invalid record format: %s", line)
	}

	// Parse start and end IPs
	startIP := ipaddr.NewIPAddressString(fields[0])
	endIP := ipaddr.NewIPAddressString(fields[1])

	startAddr, err := startIP.ToAddress()
	if err != nil {
		return fmt.Errorf("error parsing start IP: %v", err)
	}

	endAddr, err := endIP.ToAddress()
	if err != nil {
		return fmt.Errorf("error parsing end IP: %v", err)
	}

	// Parse ASN (removing any "AS" prefix if present)
	asnStr := fields[2]
	asn, err2 := strconv.Atoi(asnStr)
	if err2 != nil {
		log.Fatal(err)
	}

	// Create ASN info
	info := &ASNInfo{
		ASN:     asn,
		Country: fields[3],
		Network: fields[4],
	}

	// Create IP range
	ipRange := ipaddr.NewSequentialRange(startAddr, endAddr)
	prefixBlocks := ipRange.SpanWithPrefixBlocks()
	for _, block := range prefixBlocks {
		fmt.Println(block)
		t.trie.Put(block, info)
	}

	return nil
}

// LookupIP finds the ASN info for a given IP address
func (t *ASNTrie) LookupIP(ipStr string) (*ASNInfo, error) {
	// Parse the IP address
	ip := ipaddr.NewIPAddressString(ipStr)
	addr, err := ip.ToAddress()
	if err != nil {
		return nil, fmt.Errorf("error parsing IP address: %v", err)
	}

	// Look up in trie
	node := t.trie.LongestPrefixMatch(addr)
	if node == nil {
		return nil, fmt.Errorf("no matching ASN record found for IP: %s", ipStr)
	}
	fmt.Println(node)
	return nil, nil
}
