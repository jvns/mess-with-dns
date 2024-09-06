package parsing

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/joeig/go-powerdns/v3"
	"github.com/miekg/dns"
	"golang.org/x/net/idna"
)

type RecordRequest struct {
	Subdomain string            `json:"subdomain"`
	Typ       string            `json:"type"`
	TTL       uint32            `json:"ttl"`
	Values    map[string]string `json:"values"`
}

type RecordResponse struct {
	Subdomain  string            `json:"subdomain"`
	Type       string            `json:"type"`
	TTL        string            `json:"ttl"`
	Content    string            `json:"content"`
	DomainName string            `json:"domain_name"`
	Values     map[string]string `json:"values"`
}

func (r *RecordResponse) MarshalJSON() ([]byte, error) {
	m := map[string]string{
		"subdomain":   r.Subdomain,
		"type":        r.Type,
		"ttl":         r.TTL,
		"content":     r.Content,
		"domain_name": r.DomainName,
	}
	for k, v := range r.Values {
		m["value_"+k] = v
	}
	return json.Marshal(m)
}

func RRsetToRecordResponse(rrset *powerdns.RRset) ([]RecordResponse, error) {
	responses := []RecordResponse{}
	subdomain := extractSubSubdomain(*rrset.Name)
	for _, record := range rrset.Records {
		resp := RecordResponse{
			Subdomain:  subdomain,
			Type:       string(*rrset.Type),
			TTL:        fmt.Sprintf("%d", *rrset.TTL),
			DomainName: *rrset.Name,
		}
		values, err := ParseValues(*record.Content, string(*rrset.Type))
		if err != nil {
			return nil, err
		}
		resp.Values = values
		resp.Content = *record.Content
		responses = append(responses, resp)
	}
	return responses, nil
}

func extractSubSubdomain(name string) string {
	parts := strings.Split(name, ".")
	if len(parts) == 4 {
		return "@"
	}
	return strings.Join(parts[0:len(parts)-4], ".")
}

func fullName(subdomain string, username string) string {
	if subdomain == "@" {
		return username + ".messwithdns.com."
	}

	return fmt.Sprintf("%s.%s.messwithdns.com.", subdomain, username)
}

func ParseRecordRequest(jsRecord map[string]string, username string) (*powerdns.RRset, error) {
	rr, err := toRecordRequest(jsRecord)
	if err != nil {
		return nil, err
	}
	return parseRecordRequest(rr, username)
}

func toRecordRequest(jsRecord map[string]string) (*RecordRequest, error) {
	// Remove leading and trailing spaces because people end up entering
	// trailing spaces a lot by accident
	for k, v := range jsRecord {
		jsRecord[k] = strings.TrimSpace(v)
	}
	subdomain, ok := jsRecord["subdomain"]
	if !ok {
		return nil, fmt.Errorf("subdomain is required")
	}
	typ, ok := jsRecord["type"]
	if !ok {
		return nil, fmt.Errorf("type is required")
	}
	ttl, ok := jsRecord["ttl"]
	if !ok {
		return nil, fmt.Errorf("ttl is required")
	}
	values := map[string]string{}
	for k, v := range jsRecord {
		if k != "subdomain" && k != "type" && k != "ttl" {
			// remove the "value_" prefix
			if !strings.HasPrefix(k, "value_") {
				return nil, fmt.Errorf("invalid key: %s", k)
			}
			values[strings.TrimPrefix(k, "value_")] = v
		}
	}

	// parse the TTL as a uint32
	ttlInt, err := strconv.ParseUint(ttl, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("Error: TTL must be a number from 1 to 2147483647, got \"%s\"", ttl)
	}

	return &RecordRequest{
		Subdomain: subdomain,
		Typ:       typ,
		TTL:       uint32(ttlInt),
		Values:    values,
	}, nil
}

func parseRecordRequest(jsRecord *RecordRequest, username string) (*powerdns.RRset, error) {
	name, err := idna.ToASCII(fullName(jsRecord.Subdomain, username))
	if err != nil {
		return nil, fmt.Errorf("failed to convert name to punycode: %s", err)
	}
	name = strings.ToLower(name)

	// check if the dns name is valid

	if _, ok := dns.IsDomainName(name); !ok {
		return nil, fmt.Errorf("invalid domain name: %s", name)
	}

	typ := powerdns.RRType(jsRecord.Typ)
	content, err := parseContent(jsRecord.Values, jsRecord.Typ)
	if err != nil {
		return nil, err
	}
	return &powerdns.RRset{
		Name:    &name,
		Type:    &typ,
		TTL:     &jsRecord.TTL,
		Records: []powerdns.Record{{Content: &content}},
	}, nil
}

func getType(typ string) (RR, error) {
	switch typ {
	case "A":
		return &A{}, nil
	case "AAAA":
		return &AAAA{}, nil
	case "CAA":
		return &CAA{}, nil
	case "CNAME":
		return &CNAME{}, nil
	case "MX":
		return &MX{}, nil
	case "NS":
		return &NS{}, nil
	case "PTR":
		return &PTR{}, nil
	case "SRV":
		return &SRV{}, nil
	case "TXT":
		return &TXT{}, nil
	case "SOA":
		return &SOA{}, nil
	case "SVCB":
		return &SVCB{}, nil
	case "HTTPS":
		// HTTPS and SVCB work the same way
		return &SVCB{}, nil
	}
	return nil, fmt.Errorf("Unsupported record type: %s", typ)
}

func parseContent(values map[string]string, typ string) (string, error) {
	r, err := getType(typ)
	if err != nil {
		return "", err
	}
	return r.ToPDNS(values)
}

func ParseValues(content string, typ string) (map[string]string, error) {
	r, err := getType(typ)
	if err != nil {
		return nil, err
	}
	return r.FromPDNS(content)
}
