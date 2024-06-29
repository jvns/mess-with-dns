package streamer_test

import (
	"testing"

	"github.com/jvns/mess-with-dns/streamer"
	"github.com/stretchr/testify/assert"
)

func TestExtract(t *testing.T) {
	assert.Equal(t, streamer.ExtractSubdomain("a.b.messwithdns.com."), "b")
	assert.Equal(t, streamer.ExtractSubdomain("test.com"), "")
}
