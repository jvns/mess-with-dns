package ip2asn_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/jvns/mess-with-dns/streamer/ip2asn"
	"github.com/stretchr/testify/assert"
)

func getRanges(t *testing.T) *ip2asn.Ranges {
	ranges, err := ip2asn.NewRanges("../../../ip-ranges.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	return ranges
}

func TestParseASNv4(t *testing.T) {
	ranges := getRanges(t)

	r, _ := ranges.FindASN(net.ParseIP("8.8.8.8"))
	assert.Equal(t, 15169, r.Num)
	assert.Equal(t, "GOOGLE", r.Name)

	_, err := ranges.FindASN(net.ParseIP("255.255.255.255"))
	assert.Equal(t, err.Error(), "not found")

	_, err = ranges.FindASN(net.ParseIP("0.0.0.0"))
	assert.Equal(t, err.Error(), "not found")

}

func TestParseASNv6(t *testing.T) {
	ranges := getRanges(t)
	r, err := ranges.FindASN(net.ParseIP("2607:f8b0:4006:824::200e"))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", r)
	assert.Equal(t, 15169, r.Num)
}
