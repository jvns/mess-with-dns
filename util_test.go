package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtract(t *testing.T) {
	assert.Equal(t, ExtractSubdomain("a.b.messwithdns.com."), "b")
	assert.Equal(t, ExtractSubdomain("test.com"), "")
}
