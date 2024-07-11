package parsing

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/miekg/dns"
	"golang.org/x/net/idna"
)

type JSRecordValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type JSRecord struct {
	Subdomain string          `json:"subdomain"`
	Typ       string          `json:"type"`
	TTL       uint32          `json:"ttl"`
	Values    []JSRecordValue `json:"values"`
}

func extractSubSubdomain(name string) string {
	parts := strings.Split(name, ".")
	if len(parts) == 4 {
		return "@"
	}
	return strings.Join(parts[0:len(parts)-4], ".")
}

func RRToJSRecord(rr dns.RR) (*JSRecord, error) {
	values, err := RRToValues(rr)
	if err != nil {
		return nil, err
	}
	return &JSRecord{
		Typ:       dns.TypeToString[rr.Header().Rrtype],
		Subdomain: extractSubSubdomain(rr.Header().Name),
		TTL:       rr.Header().Ttl,
		Values:    values,
	}, nil
}

func checkValid(rr dns.RR) (dns.RR, error) {
	// make sure we have a valid RR
	// this prevents problems like invalid FQDNs in a record's fields
	msg := make([]byte, dns.Len(rr))
	_, err := dns.PackRR(rr, msg, 0, nil, false)
	if err != nil {
		return nil, fmt.Errorf("Invalid RR: %s, %#v", err, rr)
	}
	return rr, nil
}

func ParseJSRecord(jsonString []byte, username string) (dns.RR, error) {
	rr, err := parseJSRecord(jsonString, username)
	if err != nil {
		return nil, err
	}
	return checkValid(rr)
}

func fullName(subdomain string, username string) string {
	if subdomain == "@" {
		return username + ".messwithdns.com."
	}

	return fmt.Sprintf("%s.%s.messwithdns.com.", subdomain, username)
}

func parseJSRecord(jsonString []byte, username string) (dns.RR, error) {
	var jsRecord JSRecord
	err := json.Unmarshal([]byte(jsonString), &jsRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json: %s", err)
	}
	name, err := idna.ToASCII(fullName(jsRecord.Subdomain, username))
	if err != nil {
		return nil, fmt.Errorf("failed to convert name to punycode: %s", err)
	}
	hdr := dns.RR_Header{
		Name:   name,
		Rrtype: dns.StringToType[jsRecord.Typ],
		Class:  dns.ClassINET,
		Ttl:    jsRecord.TTL,
	}
	return parseValues(jsRecord.Values, hdr)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func toMap(values []JSRecordValue) map[string]string {
	m := make(map[string]string)
	for _, v := range values {
		m[v.Name] = v.Value
	}
	return m
}

type ValueMissingError struct {
	Name string
}

func (e *ValueMissingError) Error() string {
	return fmt.Sprintf("Value missing from map: %s", e.Name)
}

type UInt16Error struct {
	Name  string
	Value string
}

func (e *UInt16Error) Error() string {
	return fmt.Sprintf("%s is not between 0-65536: %s", e.Name, e.Value)
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

func has(m map[string]string, key string) bool {
	_, ok := m[key]
	return ok
}

func parseValues(values []JSRecordValue, hdr dns.RR_Header) (dns.RR, error) {
	m := toMap(values)
	switch hdr.Rrtype {
	case dns.TypeTXT:
		txt, ok := m["Txt"]
		if !ok {
			return nil, &ValueMissingError{"Txt"}
		}
		txtarray := []string{}
		for i := 0; i < len(txt); i += 255 {
			txtarray = append(txtarray, txt[i:min(i+255, len(txt))])
		}
		return &dns.TXT{Txt: txtarray, Hdr: hdr}, nil
	case dns.TypeCNAME:
		target, ok := m["Target"]
		if !ok {
			return nil, &ValueMissingError{"Target"}
		}
		return &dns.CNAME{Target: dns.Fqdn(target), Hdr: hdr}, nil
	case dns.TypeA:
		ip, ok := m["A"]
		if !ok {
			return nil, &ValueMissingError{"A"}
		}
		parsed := net.ParseIP(ip)
		if parsed == nil {
			return nil, fmt.Errorf("Invalid IP: %s", ip)
		}
		return &dns.A{A: parsed, Hdr: hdr}, nil
	case dns.TypeAAAA:
		ip, ok := m["AAAA"]
		if !ok {
			return nil, &ValueMissingError{"AAAA"}
		}
		parsed := net.ParseIP(ip)
		if parsed == nil {
			return nil, fmt.Errorf("Invalid IP: %s", ip)
		}
		return &dns.AAAA{AAAA: parsed, Hdr: hdr}, nil
	case dns.TypeNS:
		ns, ok := m["Ns"]
		if !ok {
			return nil, &ValueMissingError{"Ns"}
		}
		return &dns.NS{Ns: dns.Fqdn(ns), Hdr: hdr}, nil
	case dns.TypeMX:
		mx, ok := m["Mx"]
		if !ok {
			return nil, &ValueMissingError{"Mx"}
		}
		preference, ok := m["Preference"]
		if !ok {
			return nil, &ValueMissingError{"Preference"}
		}
		pref, err := toUint16(preference)
		if err != nil {
			return nil, &UInt16Error{"Preference", preference}
		}
		return &dns.MX{Mx: dns.Fqdn(mx), Preference: pref, Hdr: hdr}, nil
	case dns.TypeSRV:
		target, ok := m["Target"]
		if !ok {
			return nil, &ValueMissingError{"Target"}
		}
		port, ok := m["Port"]
		if !ok {
			return nil, &ValueMissingError{"Port"}
		}
		weight, ok := m["Weight"]
		if !ok {
			return nil, &ValueMissingError{"Weight"}
		}
		priority, ok := m["Priority"]
		if !ok {
			return nil, &ValueMissingError{"Priority"}
		}
		weight_uint, err := toUint16(weight)
		if err != nil {
			return nil, &UInt16Error{"Weight", weight}
		}
		priority_uint, err := toUint16(priority)
		if err != nil {
			return nil, &UInt16Error{"Priority", priority}
		}
		port_uint, err := toUint16(port)
		if err != nil {
			return nil, &UInt16Error{"Port", port}
		}
		return &dns.SRV{Target: dns.Fqdn(target), Port: port_uint, Weight: weight_uint, Priority: priority_uint, Hdr: hdr}, nil
	case dns.TypeCAA:
		flag, ok := m["Flag"]
		if !ok {
			return nil, &ValueMissingError{"Flag"}
		}
		tag, ok := m["Tag"]
		if !ok {
			return nil, &ValueMissingError{"Tag"}
		}
		value, ok := m["Value"]
		if !ok {
			return nil, &ValueMissingError{"Value"}

		}
		flag_uint, err := strconv.ParseUint(flag, 10, 8)
		if err != nil {
			return nil, &UInt8Error{"Flag", flag}
		}
		return &dns.CAA{Flag: uint8(flag_uint), Tag: tag, Value: value, Hdr: hdr}, nil
	case dns.TypePTR:
		ptr, ok := m["Ptr"]
		if !ok {
			return nil, &ValueMissingError{"Ptr"}
		}
		return &dns.PTR{Ptr: dns.Fqdn(ptr), Hdr: hdr}, nil
	}
	return nil, fmt.Errorf("Unsupported record type: %s", dns.TypeToString[hdr.Rrtype])
}

func RRToValues(rr dns.RR) ([]JSRecordValue, error) {
	switch rr := rr.(type) {
	case *dns.A:
		return []JSRecordValue{{Name: "A", Value: rr.A.String()}}, nil
	case *dns.AAAA:
		return []JSRecordValue{{Name: "AAAA", Value: rr.AAAA.String()}}, nil
	case *dns.CAA:
		return []JSRecordValue{{Name: "Flag", Value: strconv.Itoa(int(rr.Flag))}, {Name: "Tag", Value: rr.Tag}, {Name: "Value", Value: rr.Value}}, nil
	case *dns.CNAME:
		return []JSRecordValue{{Name: "Target", Value: dns.Fqdn(rr.Target)}}, nil
	case *dns.MX:
		return []JSRecordValue{{Name: "Preference", Value: strconv.Itoa(int(rr.Preference))}, {Name: "Mx", Value: dns.Fqdn(rr.Mx)}}, nil
	case *dns.NS:
		return []JSRecordValue{{Name: "Ns", Value: dns.Fqdn(rr.Ns)}}, nil
	case *dns.SRV:
		return []JSRecordValue{{Name: "Priority", Value: strconv.Itoa(int(rr.Priority))}, {Name: "Weight", Value: strconv.Itoa(int(rr.Weight))}, {Name: "Port", Value: strconv.Itoa(int(rr.Port))}, {Name: "Target", Value: dns.Fqdn(rr.Target)}}, nil
	case *dns.PTR:
		return []JSRecordValue{{Name: "Ptr", Value: dns.Fqdn(rr.Ptr)}}, nil
	case *dns.TXT:
		return []JSRecordValue{{Name: "Txt", Value: strings.Join(rr.Txt, "")}}, nil
	}
	return nil, fmt.Errorf("Unsupported record type: %s", dns.TypeToString[rr.Header().Rrtype])
}
