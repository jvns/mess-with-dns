package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	err := validateDomainName("www.messwithdns.com", "www")
	assert.NotNil(t, err, "must be fully qualified")

	err = validateDomainName("www.messwithdns.com.", "www")
	assert.NotNil(t, err, "www is invalid")

	err = validateDomainName("test.a.b.www.messwithdns.com.", "www")
	assert.NotNil(t, err, "www is invalid")

	err = validateDomainName("asdf.messwithdns.com.asdf.messwithdns.com.", "asdf")
	assert.NotNil(t, err, "messwithdns occurs twice")

	err = validateDomainName("x..messwithdns.com.", "asdf")
	assert.NotNil(t, err, "invalid domain name")

	err = validateDomainName("asdf.test.messwithdns.com.", "test")
	assert.Nil(t, err)

	err = validateDomainName("a.b.c.d.messwithdns.com.", "d")
	assert.Nil(t, err)
}

func TestGetSubdomain(t *testing.T) {
	subdomain := getSubdomain("www.messwithdns.com.")
	assert.Equal(t, "www", subdomain)

	subdomain = getSubdomain("a.b.messwithdns.com.")
	assert.Equal(t, "b", subdomain)

	subdomain = getSubdomain("messwithdns.com.")
	assert.Equal(t, "", subdomain)

	subdomain = getSubdomain("bananas.com.")
	assert.Equal(t, "", subdomain)

}
