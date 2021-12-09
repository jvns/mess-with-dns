package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	err := validateDomainName("www.messwithdns.net", "www")
	assert.NotNil(t, err, "must be fully qualified")

	err = validateDomainName("www.messwithdns.net.", "www")
	assert.NotNil(t, err, "www is invalid")

	err = validateDomainName("test.a.b.www.messwithdns.net.", "www")
	assert.NotNil(t, err, "www is invalid")

	err = validateDomainName("asdf.messwithdns.net.asdf.messwithdns.net.", "asdf")
	assert.NotNil(t, err, "messwithdns occurs twice")

	err = validateDomainName("x..messwithdns.net.", "asdf")
	assert.NotNil(t, err, "invalid domain name")

	err = validateDomainName("asdf.test.messwithdns.net.", "test")
	assert.Nil(t, err)

	err = validateDomainName("a.b.c.d.messwithdns.net.", "d")
	assert.Nil(t, err)
}

func TestGetSubdomain(t *testing.T) {
	subdomain := getSubdomain("www.messwithdns.net.")
	assert.Equal(t, "www", subdomain)

	subdomain = getSubdomain("a.b.messwithdns.net.")
	assert.Equal(t, "b", subdomain)

	subdomain = getSubdomain("messwithdns.net.")
	assert.Equal(t, "", subdomain)

	subdomain = getSubdomain("bananas.com.")
	assert.Equal(t, "", subdomain)

}
