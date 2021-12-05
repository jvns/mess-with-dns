package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	err := subdomainError("www.messwithdns.com")
	assert.NotNil(t, err, "must be fully qualified")

	err = subdomainError("www.messwithdns.com.")
	assert.NotNil(t, err, "www is invalid")

	err = subdomainError("test.a.b.www.messwithdns.com.")
	assert.NotNil(t, err, "www is invalid")

	err = subdomainError("asdf.messwithdns.com.asdf.messwithdns.com.")
	assert.NotNil(t, err, "messwithdns occurs twice")

	err = subdomainError("x..messwithdns.com.")
	assert.NotNil(t, err, "invalid domain name")

	err = subdomainError("asdf.test.messwithdns.com.")
	assert.Nil(t, err)

	err = subdomainError("a.b.c.d.messwithdns.com.")
	assert.Nil(t, err)
}
