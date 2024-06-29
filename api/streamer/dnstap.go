package streamer

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	dnstap "github.com/dnstap/golang-dnstap"
	"github.com/jvns/mess-with-dns/db"
	"github.com/jvns/mess-with-dns/streamer/ip2asn"
	"github.com/miekg/dns"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/proto"
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

	go logger.run(ctx, dnstapAddress)
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

func (l *Logger) run(ctx context.Context, host string) {
	defer l.db.Close()
	listener, err := net.Listen("tcp", host)
	if err != nil {
		panic(fmt.Sprintf("Error opening dnstap listener: %s", err.Error()))
	}
	defer listener.Close()
	input := dnstap.NewFrameStreamSockInput(listener)
	c := make(chan []byte)
	go input.ReadInto(c)

	for {
		select {
		case <-ctx.Done():
			return
		case frame := <-c:
			tap := new(dnstap.Dnstap)
			err = proto.Unmarshal(frame, tap)
			if err != nil {
				log.Printf("could not decode dnstap message: %v\n", err)
				continue
			}
			err = l.logMessage(tap.Message)
			if err != nil {
				log.Printf("could not log message: %v\n", err)
			}
		}
	}
}

func (l *Logger) logMessage(msg *dnstap.Message) error {
	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "dns.request")

	remote_addr := net.IP(msg.QueryAddress)
	remote_host := lookupHost(ctx, l.ipRanges, remote_addr)

	// TODO: why don't we have a query message here?????
	// also do we even care?? unclear

	resp := new(dns.Msg)
	err := resp.Unpack(msg.ResponseMessage)
	if err != nil {
		return fmt.Errorf("could not unpack response message: %v", err)
	}

	span.SetAttributes(attribute.String("dns.remote_addr", remote_addr.String()))
	span.SetAttributes(attribute.String("dns.remote_host", remote_host))
	span.SetAttributes(attribute.Int("dns.answer_count", len(resp.Answer)))

	err = l.LogRequest(ctx, resp, remote_addr, remote_host)
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
