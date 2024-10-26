package streamer

import (
	"context"
	"database/sql"
	"fmt"
	"net"

	"github.com/jvns/mess-with-dns/streamer/ip2asn"
	"github.com/miekg/dns"
	"go.opentelemetry.io/otel/attribute"
)

// func NewReader(r io.Reader, opt *ReaderOptions) (Reader, error)

type Logger struct {
	ipRanges *ip2asn.Ranges
	db       *sql.DB
}

func Init(ctx context.Context, workdir string, requestDBFilename string, ipRangesDBFilename string) (*Logger, error) {
	ranges, err := ip2asn.NewRanges(ipRangesDBFilename)
	if err != nil {
		return nil, fmt.Errorf("could not read ranges: %v", err)
	}

	ldb, err := connectDB(requestDBFilename)
	if err != nil {
		return nil, fmt.Errorf("could not connect to db: %v", err)
	}

	logger := &Logger{
		ipRanges: ranges,
		db:       ldb,
	}

	return logger, nil
}

func getIP(w dns.ResponseWriter) (net.IP, error) {
	if addr, ok := w.RemoteAddr().(*net.TCPAddr); ok {
		return addr.IP, nil
	} else if addr, ok := w.RemoteAddr().(*net.UDPAddr); ok {
		return addr.IP, nil
	}
	return nil, fmt.Errorf("Needs to be a TCP or UDP address")
}

func (l *Logger) Log(resp *dns.Msg, w dns.ResponseWriter) error {
	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "dns.request")

	remote_addr, err := getIP(w)
	if err != nil {
		return err
	}
	remote_host := lookupHost(ctx, l.ipRanges, remote_addr)

	// TODO: should we put the query message back? do we care? idk.

	span.SetAttributes(attribute.String("dns.remote_addr", remote_addr.String()))
	span.SetAttributes(attribute.String("dns.remote_host", remote_host))
	span.SetAttributes(attribute.Int("dns.answer_count", len(resp.Answer)))

	err = l.logRequest(ctx, resp, remote_addr, remote_host)
	if err != nil {
		return fmt.Errorf("could not log request: %v", err)
	}
	return nil
}

func lookupHost(ctx context.Context, ranges *ip2asn.Ranges, host net.IP) string {
	_, span := tracer.Start(ctx, "lookupHost")
	span.SetAttributes(attribute.String("host", host.String()))
	defer span.End()
	//don't do reverse DNS lookup, it's slow
	//names, err := net.LookupAddr(host.String())
	//if err == nil && len(names) > 0 {
	//	return names[0]
	//}
	// otherwise search ASN database
	r, err := ranges.FindASN(host)
	if err != nil {
		return ""
	}
	return r.Name
}
