package main

import (
	"fmt"
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

func validateDomainName(domain string, username string) error {
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
	subdomain := ExtractSubdomain(domain)
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
