package main

import (
	"os"
	"testing"
)

func TestSecureCookie(t *testing.T) {
	// not the real key
	base64Hash := "/JLayjTcQf0wl/YifN7WqyP6U1+y/qnxxNzhbQ1Falk="
	base64Block := "SaJ+upj49i3BzLP46bUh5g860DgB+V5z4zuTlevI9ug="
	os.Setenv("HASH_KEY", base64Hash)
	os.Setenv("BLOCK_KEY", base64Block)
	getSecureCookie()
}
