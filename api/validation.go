package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/miekg/dns"
)

var disallowedDomains = map[string]bool{
	"ns1":    true,
	"ns2":    true,
	"orange": true,
	"purple": true,
	"www":    true,
}

func subdomainError(domain string) error {
	if !strings.HasSuffix(domain, ".") {
		return fmt.Errorf("Domain must end with a period")
	}
	if _, ok := dns.IsDomainName(domain); !ok {
		return fmt.Errorf("Invalid domain name: %s", domain)
	}
	if !strings.HasSuffix(domain, ".messwithdns.com.") {
		return fmt.Errorf("Subdomain must end with .messwithdns.com.")
	}
	name := strings.TrimSuffix(domain, ".messwithdns.com.")
	// get last component of domain
	parts := strings.Split(name, ".")
	subdomain := strings.ToLower(parts[len(parts)-1])
	if _, ok := disallowedDomains[subdomain]; ok {
		return fmt.Errorf("Subdomain %s is not allowed", subdomain)
	}
	if strings.Contains(name, "messwithdns") {
		return fmt.Errorf("You tried to create a record for %s, you probably didn't want that.", domain)
	}
	return nil
}

func validateSubdomain(name string, w http.ResponseWriter) bool {
	if err := subdomainError(name); err != nil {
		fmt.Println("Error validating subdomain: ", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return false
	}
	return true
}
