package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWord(t *testing.T) {
	x := randomWord()
	assert.NotEqual(t, "", x)
}

func TestNewSubdomain(t *testing.T) {
	existing := []string{"test4", "test5", "test7"}

	x := smallestMissing("test", existing)
	assert.Equal(t, "test6", x)
}
