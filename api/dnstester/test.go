package dnstester

import (
	"github.com/miekg/dns"
	"net"
)

// DNSHandler is an interface that matches the dns.Handler interface
type DNSHandler interface {
	ServeDNS(w dns.ResponseWriter, r *dns.Msg)
}

// FakeDNSWriter implements dns.ResponseWriter for testing
type FakeDNSWriter struct {
	WrittenMsg *dns.Msg
}

func (f *FakeDNSWriter) LocalAddr() net.Addr { return nil }
func (f *FakeDNSWriter) RemoteAddr() net.Addr {
	return &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 53,
	}
}
func (f *FakeDNSWriter) WriteMsg(m *dns.Msg) error { f.WrittenMsg = m; return nil }
func (f *FakeDNSWriter) Write([]byte) (int, error) { return 0, nil }
func (f *FakeDNSWriter) Close() error              { return nil }
func (f *FakeDNSWriter) TsigStatus() error         { return nil }
func (f *FakeDNSWriter) TsigTimersOnly(bool)       {}
func (f *FakeDNSWriter) Hijack()                   {}

// DNSTester is the main struct for testing DNS handlers
type DNSTester struct {
	Handler DNSHandler
}

// NewDNSTester creates a new DNSTester with the given handler
func NewDNSTester(handler DNSHandler) *DNSTester {
	return &DNSTester{Handler: handler}
}

// ServeDNS processes a DNS request and returns the response
func (t *DNSTester) Request(request *dns.Msg) *dns.Msg {
	writer := &FakeDNSWriter{}
	t.Handler.ServeDNS(writer, request)
	return writer.WrittenMsg
}
