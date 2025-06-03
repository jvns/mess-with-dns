package streamer

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/base64"
	"net"

	"github.com/miekg/dns"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	_ "modernc.org/sqlite"
)

//go:embed create.sql
var create_sql string

func connectDB(dbFile string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	_, err = db.Exec(create_sql)
	if err != nil {
		return nil, err
	}
	return db, nil
}

var tracer = otel.Tracer("main")

func (logger Logger) DeleteOldRequests(ctx context.Context) error {
	_, span := tracer.Start(ctx, "db.DeleteOldRequests")
	defer span.End()
	// delete requests where created_at timestamp is more than a day ago
	_, err := logger.db.Exec("DELETE FROM dns_requests WHERE created_at < (strftime('%s', 'now') - (2 * 24 * 60 * 60));")
	if err != nil {
		return err
	}
	return nil
}

func serializeMsg(msg *dns.Msg) (string, error) {
	// Convert to wire format (binary)
	wire, err := msg.Pack()
	if err != nil {
		return "", err
	}

	// Convert binary to base64
	encoded := base64.StdEncoding.EncodeToString(wire)

	return encoded, nil
}

func deserializeMsg(encoded string) (*dns.Msg, error) {
	// Decode base64 to binary
	wire, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	// Unpack binary to dns.Msg
	msg := new(dns.Msg)
	err = msg.Unpack(wire)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (l *Logger) logRequest(ctx context.Context, response *dns.Msg, src_ip net.IP, src_host string) error {
	_, span := tracer.Start(ctx, "db.LogRequest")
	defer span.End()
	serializedResp, err := serializeMsg(response)
	if err != nil {
		return err
	}
	name := response.Question[0].Name
	subdomain := ExtractSubdomain(name)
	err = writeToStreams(subdomain, response, src_host, src_ip)
	if err != nil {
		return err
	}
	_, err = l.db.Exec("INSERT INTO dns_requests (name, subdomain,response, src_ip, src_host) VALUES ($1, $2, $3, $4, $5)", name, subdomain, serializedResp, src_ip.String(), src_host)
	if err != nil {
		return err
	}

	return nil
}

func (l *Logger) DeleteRequestsForDomain(ctx context.Context, subdomain string) error {
	_, span := tracer.Start(ctx, "db.DeleteRequestsForDomain")
	span.SetAttributes(attribute.String("subdomain", subdomain))
	defer span.End()
	_, err := l.db.Exec("DELETE FROM dns_requests WHERE subdomain = $1", subdomain)
	if err != nil {
		return err
	}
	return nil
}
func (l *Logger) GetRequests(ctx context.Context, subdomain string) ([]StreamLog, error) {
	_, span := tracer.Start(ctx, "db.GetRequests")
	span.SetAttributes(attribute.String("subdomain", subdomain))
	defer span.End()
	rows, err := l.db.Query("SELECT id, created_at, response, src_ip, src_host FROM dns_requests WHERE subdomain = $1 ORDER BY created_at DESC LIMIT 100", subdomain)
	if err != nil {
		return nil, err
	}
	logs := []StreamLog{}
	for rows.Next() {
		var id int
		var created_at int32
		var response []byte
		var src_ip string
		var src_host string

		err = rows.Scan(&id, &created_at, &response, &src_ip, &src_host)
		if err != nil {
			return nil, err
		}
		msg, err := deserializeMsg(string(response))
		if err != nil {
			// TODO: do we need to worry about this?
			continue
		}
		log := responseToStreamLog(int64(created_at), msg, src_host, src_ip)
		logs = append(logs, log)
	}
	return logs, nil
}
