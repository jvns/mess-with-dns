package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/miekg/dns"
)

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
	schemas := map[string][]map[string]string{}

	schemas["A"] = genSchema(dns.A{}, map[string]string{"A": "IPv4 Address"})
	schemas["AAAA"] = genSchema(dns.AAAA{}, map[string]string{"AAAA": "IPv6 Address"})
	schemas["CAA"] = genSchema(dns.CAA{}, map[string]string{"Value": "CA domain name"})
	schemas["CERT"] = genSchema(dns.CERT{}, map[string]string{"Type": "Cert type", "KeyTag": "Key tag"})
	schemas["CNAME"] = genSchema(dns.CNAME{}, map[string]string{"Cname": "Canonical Name"})
	schemas["DS"] = genSchema(dns.DS{}, map[string]string{"KeyTag": "Key tag", "Algorithm": "Algorithm", "DigestType": "Digest type"})
	schemas["MX"] = genSchema(dns.MX{}, map[string]string{"Mx": "Mail Server"})
	schemas["NS"] = genSchema(dns.NS{}, map[string]string{"Ns": "Nameserver "})
	schemas["PTR"] = genSchema(dns.PTR{}, map[string]string{"Ptr": "Pointer"})
	schemas["SOA"] = genSchema(dns.SOA{}, map[string]string{"Minttl": "Minimum TTL", "Ns": "Name Server", "Mbox": "Email address"})
	schemas["SRV"] = genSchema(dns.SRV{}, map[string]string{"Srv": "Service"})
	schemas["SRV"] = genSchema(dns.SRV{}, map[string]string{})
	schemas["TXT"] = genSchema(dns.TXT{}, map[string]string{"Txt": "Content"})
	schemas["URI"] = genSchema(dns.URI{}, map[string]string{})

	// serialize schemas to json
	x, _ := json.MarshalIndent(schemas, "", "  ")
	// pretty print json
	fmt.Println("const schemas = " + string(x) + ";")

	rrTypes := make(map[string]uint16)
	// iterate over map
	for k, v := range dns.TypeToString {
		rrTypes[v] = k
	}
	x, _ = json.MarshalIndent(rrTypes, "", "  ")
	fmt.Println("const rrTypes = " + string(x) + ";")
}

func genSchema(x interface{}, labels map[string]string) []map[string]string {
	// use reflection to get fields of x
	// and generate schema for each field
	// and print it
	reflectType := reflect.TypeOf(x)

	// make array of maps
	schemas := make([]map[string]string, 0)

	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)
		if field.Name == "Hdr" {
			continue
		}
		schema := make(map[string]string)
		schema["name"] = field.Name
		if label, ok := labels[field.Name]; ok {
			schema["label"] = label
		} else {
			schema["label"] = field.Name
		}
		schema["type"] = getType(field.Type, field.Name)
		schema["validation"] = getValidation(field)
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
		return `matches:/^([a-zA-Z0-9-]+\.)*[a-zA-Z0-9]+\.[a-zA-Z]+$/`
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
