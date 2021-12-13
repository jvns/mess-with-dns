package main

import "strings"

func ExtractSubdomain(name string) string {
	if !strings.HasSuffix(name, ".messwithdns.com.") {
		return ""
	}
	name = strings.TrimSuffix(name, ".messwithdns.com.")
	parts := strings.Split(name, ".")
	return parts[len(parts)-1]
}
