package parsing

import (
	"fmt"
	"github.com/miekg/dns"
	"net"
	"strconv"
	"strings"
)

type ValueMissingError struct {
	Name string
}

func (e *ValueMissingError) Error() string {
	return fmt.Sprintf("Value missing from map: %s", e.Name)
}

type Uint32Error struct {
	Name  string
	Value string
}

func (e *Uint32Error) Error() string {
	return fmt.Sprintf("%s is not between 0-4294967295: %s", e.Name, e.Value)
}

type UInt16Error struct {
	Name  string
	Value string
}

func (e *UInt16Error) Error() string {
	return fmt.Sprintf("%s is not between 0-65535: %s", e.Name, e.Value)
}

type UInt8Error struct {
	Name  string
	Value string
}

func (e *UInt8Error) Error() string {
	return fmt.Sprintf("%s is not between 0-255: %s", e.Name, e.Value)
}

func toUint8(value string) (uint8, error) {
	v, err := strconv.ParseUint(value, 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(v), nil
}

func toUint16(value string) (uint16, error) {
	v, err := strconv.ParseUint(value, 10, 16)
	if err != nil {
		return 0, err
	}

	return uint16(v), nil
}

type RR interface {
	ToPDNS(m map[string]string) (string, error)
	FromPDNS(s string) (map[string]string, error)
}

type A struct{}

func getIPv4(m map[string]string, key string) (string, error) {
	addr, ok := m[key]
	if !ok {
		return "", &ValueMissingError{Name: key}
	}
	parseIP := net.ParseIP(addr)
	if parseIP == nil || parseIP.To4() == nil {
		return "", fmt.Errorf("Invalid IPv4 address: %s", addr)
	}
	return addr, nil
}

func getIPv6(m map[string]string, key string) (string, error) {
	addr, ok := m[key]
	if !ok {
		return "", &ValueMissingError{Name: key}
	}
	parseIP := net.ParseIP(addr)
	if parseIP == nil || parseIP.To16() == nil {
		return "", fmt.Errorf("Invalid IPv6 address: %s", addr)
	}
	return addr, nil
}

func getUint8(m map[string]string, key string) (uint8, error) {
	val, ok := m[key]
	if !ok {
		return 0, &ValueMissingError{Name: key}
	}
	v, err := toUint8(val)
	if err != nil {
		return 0, &UInt8Error{Name: key, Value: val}
	}
	return v, nil
}

func getUint16(m map[string]string, key string) (uint16, error) {
	val, ok := m[key]
	if !ok {
		return 0, &ValueMissingError{Name: key}
	}
	v, err := toUint16(val)
	if err != nil {
		return 0, &UInt16Error{Name: key, Value: val}
	}
	return v, nil
}

func getUint32(m map[string]string, key string) (uint32, error) {
	val, ok := m[key]
	if !ok {
		return 0, &ValueMissingError{Name: key}
	}
	v, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		return 0, &Uint32Error{Name: key, Value: val}
	}
	return uint32(v), nil
}

func getFqdn(m map[string]string, key string) (string, error) {
	fqdn, ok := m[key]
	if !ok {
		return "", &ValueMissingError{Name: key}
	}
	if _, ok := dns.IsDomainName(fqdn); !ok {
		return "", fmt.Errorf("invalid domain name: %s", fqdn)
	}
	// check if it is a valid fqdn
	return dns.Fqdn(fqdn), nil
}

func (a *A) ToPDNS(m map[string]string) (string, error) {
	addr, err := getIPv4(m, "A")
	if err != nil {
		return "", err
	}
	return addr, nil
}

func (a *A) FromPDNS(s string) (map[string]string, error) {
	return map[string]string{"A": s}, nil
}

type AAAA struct{}

func (a *AAAA) ToPDNS(m map[string]string) (string, error) {
	addr, err := getIPv6(m, "AAAA")
	if err != nil {
		return "", err
	}

	return addr, nil
}

func (a *AAAA) FromPDNS(s string) (map[string]string, error) {
	return map[string]string{"AAAA": s}, nil
}

type CNAME struct{}

func (c *CNAME) ToPDNS(m map[string]string) (string, error) {
	cname, err := getFqdn(m, "Target")
	if err != nil {
		return "", err
	}
	return cname, nil
}

func (c *CNAME) FromPDNS(s string) (map[string]string, error) {
	return map[string]string{"Target": s}, nil
}

// mx

type MX struct{}

func (r *MX) ToPDNS(m map[string]string) (string, error) {
	pref, err := getUint16(m, "Preference")
	if err != nil {
		return "", err
	}
	mx, err := getFqdn(m, "Mx")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d %s", pref, mx), nil
}

func (r *MX) FromPDNS(s string) (map[string]string, error) {
	// split string on space
	parts := strings.Split(s, " ")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid MX record: %s", s)
	}
	return map[string]string{"Preference": parts[0], "Mx": parts[1]}, nil
}

type NS struct{}

func (r *NS) ToPDNS(m map[string]string) (string, error) {
	ns, err := getFqdn(m, "Ns")
	if err != nil {
		return "", err
	}
	return ns, nil
}

func (r *NS) FromPDNS(s string) (map[string]string, error) {
	return map[string]string{"Ns": s}, nil
}

type TXT struct{}

func (r *TXT) ToPDNS(m map[string]string) (string, error) {
	txt, ok := m["Txt"]
	if !ok {
		return "", &ValueMissingError{Name: "Txt"}
	}
	return fmt.Sprintf("\"%s\"", txt), nil
}

func (r *TXT) FromPDNS(s string) (map[string]string, error) {
	s = strings.Trim(s, "\"")
	return map[string]string{"Txt": s}, nil
}

// ptr

type PTR struct{}

func (r *PTR) ToPDNS(m map[string]string) (string, error) {
	ptr, err := getFqdn(m, "Ptr")
	if err != nil {
		return "", err
	}
	return ptr, nil
}

func (r *PTR) FromPDNS(s string) (map[string]string, error) {
	return map[string]string{"Ptr": s}, nil
}

type SRV struct{}

func (r *SRV) ToPDNS(m map[string]string) (string, error) {
	priority, err := getUint16(m, "Priority")
	if err != nil {
		return "", err
	}
	weight, err := getUint16(m, "Weight")
	if err != nil {
		return "", err
	}
	port, err := getUint16(m, "Port")
	if err != nil {
		return "", err
	}
	target, err := getFqdn(m, "Target")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d %d %d %s", priority, weight, port, target), nil
}

func (r *SRV) FromPDNS(s string) (map[string]string, error) {
	parts := strings.Split(s, " ")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid SRV record: %s", s)
	}
	return map[string]string{"Priority": parts[0], "Weight": parts[1], "Port": parts[2], "Target": parts[3]}, nil
}

type CAA struct{}

func (r *CAA) ToPDNS(m map[string]string) (string, error) {
	flag, err := getUint8(m, "Flag")
	if err != nil {
		return "", err
	}
	tag, ok := m["Tag"]
	if !ok {
		return "", &ValueMissingError{Name: "Tag"}
	}
	value, ok := m["Value"]
	if !ok {
		return "", &ValueMissingError{Name: "Value"}
	}
	return fmt.Sprintf("%d %s \"%s\"", flag, tag, value), nil
}

func (r *CAA) FromPDNS(s string) (map[string]string, error) {
	parts := strings.Split(s, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid CAA record: %s", s)
	}
	return map[string]string{"Flag": parts[0], "Tag": parts[1], "Value": parts[2]}, nil
}

type SOA struct{}

func (r *SOA) ToPDNS(m map[string]string) (string, error) {
	mname, err := getFqdn(m, "Mname")
	if err != nil {
		return "", err
	}
	rname, err := getFqdn(m, "Rname")
	if err != nil {
		return "", err
	}
	serial, err := getUint32(m, "Serial")
	if err != nil {
		return "", err
	}
	refresh, err := getUint32(m, "Refresh")
	if err != nil {
		return "", err
	}
	retry, err := getUint32(m, "Retry")
	if err != nil {
		return "", err
	}
	expire, err := getUint32(m, "Expire")
	if err != nil {
		return "", err
	}
	minimum, err := getUint32(m, "Minimum")
	if err != nil {
		return "", err
	}
	// replace the first "@" with a "." in the rname
	rname = strings.Replace(rname, "@", ".", 1)
	return fmt.Sprintf("%s %s %d %d %d %d %d", mname, rname, serial, refresh, retry, expire, minimum), nil
}

func (r *SOA) FromPDNS(s string) (map[string]string, error) {
	parts := strings.Split(s, " ")
	if len(parts) != 7 {
		return nil, fmt.Errorf("invalid SOA record: %s", s)
	}
	// replace the first "." with a "@" in the rname
	parts[1] = strings.Replace(parts[1], ".", "@", 1)
	return map[string]string{"Mname": parts[0], "Rname": parts[1], "Serial": parts[2], "Refresh": parts[3], "Retry": parts[4], "Expire": parts[5], "Minimum": parts[6]}, nil
}
