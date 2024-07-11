package main

import (
	"encoding/json"
	"fmt"
	"github.com/jvns/mess-with-dns/parsing"
)

func main() {
	schemas, err := parsing.GenerateSchemas()
	if err != nil {
		fmt.Printf("Error generating schemas: %v\n", err)
		return
	}
	jsonSchemas, _ := json.MarshalIndent(schemas, "", "  ")
	fmt.Println(string(jsonSchemas))
}
