package users

import (
	"testing"
)

func TestSecureCookie(t *testing.T) {
	// not the real key
	base64Hash := "/JLayjTcQf0wl/YifN7WqyP6U1+y/qnxxNzhbQ1Falk="
	base64Block := "SaJ+upj49i3BzLP46bUh5g860DgB+V5z4zuTlevI9ug="
	us, err := Init(":memory:", base64Hash, base64Block)
	if err != nil {
		t.Fatal(err)
	}
	us.getSecureCookie()
}
