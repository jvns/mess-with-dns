package users

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

func TestNewUser(t *testing.T) {
	hash := "/JLayjTcQf0wl/YifN7WqyP6U1+y/qnxxNzhbQ1Falk="
	block := "SaJ+upj49i3BzLP46bUh5g860DgB+V5z4zuTlevI9ug="
	us, err := Init(":memory:", hash, block)
	if err != nil {
		t.Fatal(err)
	}
	_, err = us.CreateAvailableSubdomain()
	if err != nil {
		t.Fatal(err)
	}
}
