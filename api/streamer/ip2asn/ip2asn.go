package ip2asn

import (
	"database/sql"
	"encoding/binary"
	"fmt"
	"net"
	"strings"

	_ "modernc.org/sqlite"
)

type Ranges struct {
	db *sql.DB
}

type IPRange struct {
	StartIP net.IP
	EndIP   net.IP
	Num     int
	Name    string
	Country string
}

func NewRanges(dbFilename string) (*Ranges, error) {
	db, err := sql.Open("sqlite", dbFilename)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	return &Ranges{db: db}, nil
}

func (ranges *Ranges) FindASN(ip net.IP) (IPRange, error) {
	if ip.To4() != nil {
		return ranges.FindASN4(ip)
	} else {
		return ranges.FindASN6(ip)
	}
}

func ip2int(ip net.IP) uint32 {
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}

func int2ip(nn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)
	return ip
}

func (ranges *Ranges) FindASN4(ip net.IP) (IPRange, error) {
	ipInt := ip2int(ip)
	rows, err := ranges.db.Query("SELECT * FROM ipv4_ranges WHERE start_ip <= ? AND end_ip >= ?", ipInt, ipInt)
	if err != nil {
		return IPRange{}, err
	}
	defer rows.Close()
	// start and end ip are ints, need to parse them maybe
	var startIP, endIP uint32
	var r IPRange
	if rows.Next() {
		err = rows.Scan(&startIP, &endIP, &r.Num, &r.Name, &r.Country)
		if err != nil {
			return IPRange{}, err
		}
		r.StartIP = int2ip(startIP)
		r.EndIP = int2ip(endIP)
		return r, nil
	}
	return IPRange{}, fmt.Errorf("not found")
}

func expandIPv6(ip net.IP) string {
	ipv6 := ip.To16()
	// Create the expanded form
	var parts []string
	for i := 0; i < 16; i += 2 {
		parts = append(parts, fmt.Sprintf("%02x%02x", ipv6[i], ipv6[i+1]))
	}

	return strings.Join(parts, ":")
}

func (ranges *Ranges) FindASN6(ip net.IP) (IPRange, error) {
	ip_str := expandIPv6(ip)
	fmt.Println(ip_str)
	rows, err := ranges.db.Query("SELECT * FROM ipv6_ranges WHERE start_ip <= ? AND end_ip >= ?", ip_str, ip_str)
	if err != nil {
		return IPRange{}, err
	}
	defer rows.Close()
	var startIP, endIP string
	var r IPRange
	if rows.Next() {
		err = rows.Scan(&startIP, &endIP, &r.Num, &r.Name, &r.Country)
		if err != nil {
			return IPRange{}, err
		}
		r.StartIP = net.ParseIP(startIP)
		r.EndIP = net.ParseIP(endIP)
		return r, nil
	}
	return IPRange{}, fmt.Errorf("no rows found")
}
