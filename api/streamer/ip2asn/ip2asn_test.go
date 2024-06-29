package ip2asn_test

import (
	"github.com/jvns/mess-with-dns/streamer/ip2asn"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestParseASN(t *testing.T) {
	ranges, err := ip2asn.ReadRanges("../../..")
	if err != nil {
		t.Fatal(err)
	}
	r, _ := ranges.FindASN(net.ParseIP("8.8.8.8"))
	assert.Equal(t, 15169, r.Num)
	assert.Equal(t, "GOOGLE", r.Name)

	_, err = ranges.FindASN(net.ParseIP("255.255.255.255"))
	assert.Equal(t, err.Error(), "not found")

	_, err = ranges.FindASN(net.ParseIP("0.0.0.0"))
	assert.Equal(t, err.Error(), "not found")
}
