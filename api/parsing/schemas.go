package parsing

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/miekg/dns"
)

type SchemaField struct {
	Label   string `json:"label"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Width   string `json:"width"`
	Example string `json:"example"`
}

type RecordSchema map[string][]SchemaField

type FieldSpec struct {
	Label   string
	Width   string
	Example string
}

func GenerateSchemas() (RecordSchema, error) {
	schemas := make(RecordSchema)

	recordTypes := map[string]interface{}{
		"A":     dns.A{},
		"AAAA":  dns.AAAA{},
		"CNAME": dns.CNAME{},
		"MX":    dns.MX{},
		"NS":    dns.NS{},
		"PTR":   dns.PTR{},
		"SRV":   dns.SRV{},
		"TXT":   dns.TXT{},
		"CAA":   CAA{},
	}

	fieldSpecs := map[string]map[string]FieldSpec{
		"A": {
			"A": {Label: "IPv4 Address", Width: "10rem", Example: "1.2.3.4"},
		},
		"AAAA": {
			"AAAA": {Label: "IPv6 Address", Width: "10rem", Example: "2001:db8::1"},
		},
		"CAA": {
			"Flag":  {Label: "Flag", Width: "10rem", Example: "1"},
			"Tag":   {Label: "Tag", Width: "10rem", Example: "1"},
			"Value": {Label: "CA domain name", Width: "10rem", Example: "TODO"},
		},
		"CNAME": {
			"Target": {Label: "Target", Width: "10rem", Example: "orange-ip.fly.dev"},
		},
		"MX": {
			"Preference": {Label: "Preference", Width: "4rem", Example: "10"},
			"Mx":         {Label: "Mail Server", Width: "4rem", Example: "mail.messagingengine.com"},
		},
		"NS": {
			"Ns": {Label: "Nameserver", Width: "4rem", Example: "ns1.example.com"},
		},
		"PTR": {
			"Ptr": {Label: "Pointer", Width: "4rem", Example: "www.example.com"},
		},
		"SRV": {
			"Priority": {Label: "Priority", Width: "4rem", Example: "10"},
			"Weight":   {Label: "Weight", Width: "4rem", Example: "TODO"},
			"Port":     {Label: "Port", Width: "6rem", Example: "8080"},
			"Target":   {Label: "Target", Width: "10rem", Example: "orange-ip.fly.dev"},
		},
		"TXT": {
			"Txt": {Label: "Content", Width: "10rem", Example: "hello world"},
		},
	}

	for typeName, recordType := range recordTypes {
		schema, err := generateSchemaForType(recordType, fieldSpecs[typeName])
		if err != nil {
			return nil, fmt.Errorf("error generating schema for %s: %w", typeName, err)
		}
		schemas[typeName] = schema
	}

	return schemas, nil
}

func generateSchemaForType(recordType interface{}, fieldSpecs map[string]FieldSpec) ([]SchemaField, error) {
	var schema []SchemaField
	t := reflect.TypeOf(recordType)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name == "Hdr" {
			continue
		}
		spec, ok := fieldSpecs[field.Name]
		if !ok {
			return nil, fmt.Errorf("missing field specification for %s", field.Name)
		}

		if spec.Label == "" || spec.Width == "" || spec.Example == "" {
			return nil, fmt.Errorf("incomplete field specification for %s", field.Name)
		}

		schemaField := SchemaField{
			Label:   spec.Label,
			Name:    field.Name,
			Type:    getFieldType(field.Type),
			Width:   spec.Width,
			Example: spec.Example,
		}
		schema = append(schema, schemaField)
	}

	return schema, nil
}

func toTitleCase(s string) string {
	return strings.Title(strings.ToLower(s))
}

func getFieldType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "text"
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "number"
	default:
		return "text"
	}
}

// Manually specified struct for CAA record
type CAA struct {
	Flag  uint8
	Tag   string
	Value string
}
