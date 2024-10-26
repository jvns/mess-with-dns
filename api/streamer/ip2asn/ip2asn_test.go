package ip2asn_test

import (
	"fmt"
	"net/netip"
	"testing"

	"github.com/jvns/mess-with-dns/streamer/ip2asn"
	"github.com/stretchr/testify/assert"
)

func parseIP(s string) netip.Addr {
	ip, err := netip.ParseAddr(s)
	if err != nil {
		panic(err)
	}
	return ip
}

func TestParseASN(t *testing.T) {
	ranges, err := ip2asn.ReadRanges("../../..")
	if err != nil {
		t.Fatal(err)
	}
	r, _ := ranges.FindASN(parseIP("8.8.8.8"))
	var asn uint32 = 15169
	assert.Equal(t, asn, r.Num)
	assert.Equal(t, "GOOGLE", r.Name)

	_, err = ranges.FindASN(parseIP("255.255.255.255"))
	assert.Equal(t, err.Error(), "not found")

	_, err = ranges.FindASN(parseIP("0.0.0.0"))
	assert.Equal(t, err.Error(), "not found")
}

func TestParseASNv6(t *testing.T) {
	ranges, err := ip2asn.ReadRanges("../../..")
	if err != nil {
		t.Fatal(err)
	}
	r, err := ranges.FindASN(parseIP("2607:f8b0:4006:824::200e"))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", r)
	var asn uint32 = 15169
	assert.Equal(t, asn, r.Num)
	assert.Equal(t, "GOOGLE", r.Name)
	assert.Equal(t, "US", r.Country)
}
