package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	err := subdomainError("www.messwithdns.com", "www")
	assert.NotNil(t, err, "must be fully qualified")

	err = subdomainError("www.messwithdns.com.", "www")
	assert.NotNil(t, err, "www is invalid")

	err = subdomainError("test.a.b.www.messwithdns.com.", "www")
	assert.NotNil(t, err, "www is invalid")

	err = subdomainError("asdf.messwithdns.com.asdf.messwithdns.com.", "asdf")
	assert.NotNil(t, err, "messwithdns occurs twice")

	err = subdomainError("x..messwithdns.com.", "asdf")
	assert.NotNil(t, err, "invalid domain name")

	err = subdomainError("asdf.test.messwithdns.com.", "test")
	assert.Nil(t, err)

	err = subdomainError("a.b.c.d.messwithdns.com.", "d")
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
