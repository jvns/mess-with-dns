package ip2asn

import (
	"bufio"
	"errors"
	"net/netip"
	"os"
	"strconv"
	"strings"
)

type Ranges struct {
	IPv4Ranges []IPRange
	IPv6Ranges []IPRange
	asnPool    *ASNPool
}

type ASNPool struct {
	asns   []ASNInfo
	lookup map[ASNInfo]uint32
}

// NewStringPool creates a new string pool
func NewASNPool() *ASNPool {
	return &ASNPool{
		asns:   make([]ASNInfo, 0),
		lookup: make(map[ASNInfo]uint32),
	}
}

// Add adds a string to the pool and returns its index
func (ap *ASNPool) Add(asn ASNInfo) uint32 {
	if idx, exists := ap.lookup[asn]; exists {
		return idx
	}
	idx := uint32(len(ap.asns))
	ap.asns = append(ap.asns, asn)
	ap.lookup[asn] = idx
	return idx
}

func (ap *ASNPool) Get(idx uint32) ASNInfo {
	return ap.asns[idx]
}

func (ap *ASNPool) RemoveLookup() {
	ap.lookup = nil
}

type ASNInfo struct {
	Country string
	Name    string
}

type IPRange struct {
	StartIP netip.Addr
	EndIP   netip.Addr
	ASN     uint32
	Idx     uint32
}

type IPRangeHydrated struct {
	StartIP netip.Addr
	EndIP   netip.Addr
	Num     uint32
	Country string
	Name    string
}

func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

func ReadRanges(workdir string) (Ranges, error) {
	pool := NewASNPool()
	ipv4Ranges, err := ReadASNs(workdir+"/ip2asn-v4.tsv", pool)
	if err != nil {
		return Ranges{}, err
	}
	ipv6Ranges, err := ReadASNs(workdir+"/ip2asn-v6.tsv", pool)
	if err != nil {
		return Ranges{}, err
	}

	pool.RemoveLookup()
	return Ranges{
		IPv4Ranges: ipv4Ranges,
		IPv6Ranges: ipv6Ranges,
		asnPool:    pool,
	}, nil
}

func (ranges Ranges) FindASN(ip netip.Addr) (IPRangeHydrated, error) {
	var r IPRange
	var err error
	if ip.Is4() {
		r, err = FindASN(ranges.IPv4Ranges, ip)
	} else {
		r, err = FindASN(ranges.IPv6Ranges, ip)
	}
	if err != nil {
		return IPRangeHydrated{}, err
	}
	asn := ranges.asnPool.Get(r.Idx)
	return IPRangeHydrated{
		StartIP: r.StartIP,
		EndIP:   r.EndIP,
		Num:     r.ASN,
		Country: asn.Country,
		Name:    asn.Name,
	}, nil

}

func FindASN(lines []IPRange, ip netip.Addr) (IPRange, error) {
	// binary search
	start := 0
	end := len(lines) - 1
	for start <= end {
		mid := (start + end) / 2
		// check if it's between StartIP and EndIP
		above_start := ip.Compare(lines[mid].StartIP) >= 0
		below_end := ip.Compare(lines[mid].EndIP) <= 0
		if above_start && below_end {
			return lines[mid], nil
		} else if !above_start {
			end = mid - 1
		} else {
			start = mid + 1
		}
	}
	return IPRange{}, errors.New("not found")
}

func ReadASNs(filename string, ap *ASNPool) ([]IPRange, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	// read lines
	scanner := bufio.NewScanner(f)
	var lines []IPRange
	for scanner.Scan() {
		line := scanner.Text()
		// split line
		fields := strings.Split(line, "\t")
		// parse fields
		name := fields[4]
		// only take part after " - "
		if strings.Contains(name, " - ") {
			name = strings.Split(name, " - ")[1]
		}
		info := ASNInfo{
			Country: fields[3],
			Name:    name,
		}
		// add to pool
		idx := ap.Add(info)
		startIP, err := netip.ParseAddr(fields[0])
		if err != nil {
			return nil, err
		}
		endIP, err := netip.ParseAddr(fields[1])
		if err != nil {
			return nil, err
		}
		IPRange := IPRange{
			StartIP: startIP,
			EndIP:   endIP,
			ASN:     uint32(parseInt(fields[2])),
			Idx:     idx,
		}
		lines = append(lines, IPRange)
	}
	return lines, nil
}
