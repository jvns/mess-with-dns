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

func getSubdomain(domain string) string {
	if !strings.HasSuffix(domain, ".messwithdns.com.") {
		return ""
	}
	name := strings.TrimSuffix(domain, "messwithdns.com.")
	parts := strings.Split(name, ".")
	if len(parts) == 1 {
		return ""
	}
	return strings.ToLower(parts[len(parts)-2])
}

func subdomainError(domain string, username string) error {
	if !strings.HasSuffix(domain, ".") {
		return fmt.Errorf("Domain must end with a period")
	}
	if _, ok := dns.IsDomainName(domain); !ok {
		return fmt.Errorf("Invalid domain name: %s", domain)
	}
	if !strings.HasSuffix(domain, ".messwithdns.com.") {
		return fmt.Errorf("Subdomain must end with .messwithdns.com.")
	}
	// get last component of domain
	name := strings.TrimSuffix(domain, ".messwithdns.com.")
	subdomain := getSubdomain(domain)
	if subdomain != username {
		return fmt.Errorf("Subdomain must be '%s'", username)
	}
	if _, ok := disallowedDomains[subdomain]; ok {
		return fmt.Errorf("Sorry, you're not allowed to make changes to '%s' :)", subdomain)
	}
	if strings.Contains(name, "messwithdns") {
		return fmt.Errorf("You tried to create a record for %s, you probably didn't want that.", domain)
	}
	return nil
}

func validateDomainName(name string, username string, w http.ResponseWriter) bool {
	if err := subdomainError(name, username); err != nil {
		fmt.Println("Error validating subdomain: ", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return false
	}
	return true
}
