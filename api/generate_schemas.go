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
	schemas["MX"] = genSchema(dns.MX{}, map[string]string{"Mx": "Mail Server"})
	schemas["NS"] = genSchema(dns.NS{}, map[string]string{"Ns": "Name Server"})
	schemas["PTR"] = genSchema(dns.PTR{}, map[string]string{"Ptr": "Pointer"})
	// serialize schemas to json
	x, _ := json.MarshalIndent(schemas, "", "  ")
	// pretty print json
	fmt.Println(string(x))

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
		schema["validation"] = getValidation(field.Type)
		//fmt.Println("'validation-messages': {'" + getValidationMessages(field.Type) + "'},")
		schemas = append(schemas, schema)
	}
	return schemas
}

func getValidation(t reflect.Type) string {
	// get json schema for the type

	switch t.String() {
	case "net.IP":
		// TODO: check if it's an ipv4 address
		return "matches:/[0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+\\/"
	case "string":
		// check json label

		return "number"
	case "uint8":
		return "between:0,255"
	case "uint16":
		return "between:0,65535"
	case "uint32":
		return "between:0,4294967295"
	default:
		// exit program with error
		fmt.Println("Error: Unsupported type: " + t.String())
		os.Exit(1)
	}
	return ""
}
