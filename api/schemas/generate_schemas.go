package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/miekg/dns"
)

var examples = map[string]string{
	"A":           "1.2.3.4",
	"AAAA":        "2001:db8::1",
	"Algorithm":   "1",
	"Certificate": "MIIGDjCCA/agAwIBAgIJAJz/8nh5oYsMA0GCSqGSIb3DQEBBQUAMIGuMQswCQYDVQQGEwJKUDEOMAwGA1UECBMF",
	"Digest":      "QmFzZTY0IGVuY29kZWQgZm9ybWF0",
	"DigestType":  "1",
	"Expire":      "1",
	"Flag":        "1",
	"KeyTag":      "1",
	"Mbox":        "example.com",
	"Minttl":      "3600",
	"Mx":          "mail.messagingengine.com",
	"Ns":          "ns1.example.com",
	"Port":        "8080",
	"Preference":  "10",
	"Priority":    "10",
	"Ptr":         "www.example.com",
	"Refresh":     "3600",
	"Retry":       "3600",
	"Serial":      "1",
	"Tag":         "1",
	"Target":      "orange-ip.fly.dev",
	"Txt":         "\"Hello World\"",
	"Type":        "1",
	"Value":       "TODO",
	"Weight":      "TODO",
	"ttl":         "60",
}

func main() {
	// Generate schemas for the dns.A struct
	// format:
	/*
	   [{
	           'name': 'A',
	           'label': 'IPv4 Address',
	           'validation': "matches:/[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+\/",
	           'validation-messages': {
	               'matches': 'Invalid IPv4 Address'
	           },
	       },
	*/
	// get fields of dns.A type
	schemas := map[string][]map[string]interface{}{}

	schemas["A"] = genSchema(dns.A{}, map[string]string{"A": "IPv4 Address"})
	schemas["AAAA"] = genSchema(dns.AAAA{}, map[string]string{"AAAA": "IPv6 Address"})
	schemas["CAA"] = genSchema(dns.CAA{}, map[string]string{"Value": "CA domain name"})
	schemas["CNAME"] = genSchema(dns.CNAME{}, map[string]string{})
	schemas["MX"] = genSchema(dns.MX{}, map[string]string{"Mx": "Mail Server"})
	schemas["NS"] = genSchema(dns.NS{}, map[string]string{"Ns": "Nameserver "})
	schemas["PTR"] = genSchema(dns.PTR{}, map[string]string{"Ptr": "Pointer"})
	schemas["SRV"] = genSchema(dns.SRV{}, map[string]string{"Srv": "Service"})
	schemas["TXT"] = genSchema(dns.TXT{}, map[string]string{"Txt": "Content"})
	//schemas["CERT"] = genSchema(dns.CERT{}, map[string]string{"Type": "Cert type", "KeyTag": "Key tag"})
	//schemas["URI"] = genSchema(dns.URI{}, map[string]string{})
	//schemas["DS"] = genSchema(dns.DS{}, map[string]string{"KeyTag": "Key tag", "Algorithm": "Algorithm", "DigestType": "Digest type"})
	//schemas["SOA"] = genSchema(dns.SOA{}, map[string]string{"Minttl": "Minimum TTL", "Ns": "Name Server", "Mbox": "Email address"})

	// add ttl field to each schema
	for k, schema := range schemas {
		schema = append(schema, map[string]interface{}{
			"name":       "ttl",
			"label":      "TTL",
			"type":       "number",
			"validation": "required",
			"validation-messages": map[string]string{
				"required": "Example: " + examples["ttl"],
			},
		})
		schemas[k] = schema
	}

	// serialize schemas to json
	x, _ := json.MarshalIndent(schemas, "", "  ")
	file, err := os.Create("../frontend/schemas.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	file.Write(x)

	rrTypes := make(map[string]uint16)
	// iterate over map
	for k, v := range dns.TypeToString {
		rrTypes[v] = k
	}
	x, _ = json.MarshalIndent(rrTypes, "", "  ")
	file, err = os.Create("../frontend/rrTypes.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	file.Write(x)

	// rcodes.json
	file, err = os.Create("../frontend/rcodes.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	x, _ = json.MarshalIndent(dns.RcodeToString, "", "  ")
	file.Write(x)
}

func genSchema(x interface{}, labels map[string]string) []map[string]interface{} {
	// use reflection to get fields of x
	// and generate schema for each field
	// and print it
	reflectType := reflect.TypeOf(x)

	// make array of maps
	schemas := make([]map[string]interface{}, 0)

	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)
		if field.Name == "Hdr" {
			continue
		}
		schema := make(map[string]interface{})
		schema["name"] = field.Name
		if label, ok := labels[field.Name]; ok {
			schema["label"] = label
		} else {
			schema["label"] = field.Name
		}
		schema["type"] = getType(field.Type, field.Name)
		validation := getValidation(field)
		schema["validation"] = validation
		valField := strings.Split(validation, ":")[0]

		messages := map[string]string{}
		messages[valField] = "Example: " + examples[field.Name]
		schema["validation-messages"] = messages
		//fmt.Println("'validation-messages': {'" + getValidationMessages(field.Type) + "'},")
		schemas = append(schemas, schema)
	}
	return schemas
}

func getType(t reflect.Type, field_name string) string {
	if field_name == "Txt" {
		return "textarea"
	}
	switch t.String() {
	case "net.IP":
		return "text"
	case "string":
		return "text"
	case "int":
		return "number"
	case "uint":
		return "number"
	case "uint16":
		return "number"
	case "uint32":
		return "number"
	case "uint8":
		return "number"
	default:
		fmt.Println("Error: Unsupported type: " + t.String())
		os.Exit(1)
	}
	return ""
}

func getValidation(field reflect.StructField) string {
	// get json schema for the type
	field_name := field.Name
	t := field.Type

	tag := field.Tag.Get("dns")

	if tag == "cdomain-name" || tag == "domain-name" {
		return `matches:/^([a-zA-Z0-9-]+\.)*[a-zA-Z0-9]+\.[a-zA-Z]+\.?$/`
	} else if tag == "base64" {
		return `matches:/^[a-zA-Z0-9\+\/\=]+$/`
	} else if tag == "hex" {
		return `matches:/^[a-fA-F0-9]+$/`
	}

	switch t.String() {
	case "net.IP":
		// TODO: check if it's an ipv4 address
		if field_name == "A" {
			return `matches:/[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+/`
		} else if field_name == "AAAA" {
			return "matches:/[0-9a-fA-F:]+/"
		}
	case "string":
		// TODO: check json label
		return "required"
	case "uint8":
		return "number|between:0,255"
	case "uint16":
		return "number|between:0,65535"
	case "uint32":
		return "number"
	case "[]string":
		return "required" // for TXT record
	default:
		// exit program with error
		fmt.Println("Error: Unsupported type: " + t.String())
		os.Exit(1)
	}
	return ""
}
