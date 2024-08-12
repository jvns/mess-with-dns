package streamer

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/jvns/mess-with-dns/db"
	"github.com/jvns/mess-with-dns/streamer/ip2asn"
	"github.com/miekg/dns"
	"go.opentelemetry.io/otel/attribute"
)

// func NewReader(r io.Reader, opt *ReaderOptions) (Reader, error)

type Logger struct {
	ipRanges *ip2asn.Ranges
	db       *db.LockedDB
}

func Init(ctx context.Context, workdir string, dbFilename string, dnstapAddress string) (*Logger, error) {
	ranges, err := ip2asn.ReadRanges(workdir)
	if err != nil {
		return nil, fmt.Errorf("could not read ranges: %v", err)
	}

	ldb, err := connectDB(dbFilename)
	if err != nil {
		return nil, fmt.Errorf("could not connect to db: %v", err)
	}

	go cleanup(ldb)

	logger := &Logger{
		ipRanges: &ranges,
		db:       ldb,
	}

	return logger, nil
}

func cleanup(db *db.LockedDB) {
	ctx := context.Background()
	_, span := tracer.Start(ctx, "cleanup")
	defer span.End()
	for {
		fmt.Println("Deleting old requests...")
		deleteOldRequests(ctx, db)
		time.Sleep(time.Minute * 15)
	}
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
	names, err := net.LookupAddr(host.String())
	if err == nil && len(names) > 0 {
		return names[0]
	}
	// otherwise search ASN database
	r, err := ranges.FindASN(host)
	if err != nil {
		return ""
	}
	return r.Name
}
