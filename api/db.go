package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/miekg/dns"
)

// connect to planetscale
func connect() (*sql.DB, error) {
	// get connection string from environment
	db, err := sql.Open("mysql", os.Getenv("PLANETSCALE_CONNECTION_STRING"))
	if err != nil {
		return nil, err
	}
	return db, nil
}

func GetSerial(db *sql.DB) (uint32, error) {
	var serial uint32
	err := db.QueryRow("SELECT serial FROM dns_serials").Scan(&serial)
	if err != nil {
		return 0, err
	}
	return serial, nil
}

func IncrementSerial(tx *sql.Tx) error {
	_, err := tx.Exec("UPDATE dns_serials SET serial = serial + 1")
	if err != nil {
		return err
	}
	// get new serial
	var serial uint32
	err = tx.QueryRow("SELECT serial FROM dns_serials").Scan(&serial)
	if err != nil {
		return err
	}
	// commit transaction
	err = tx.Commit()
	if err != nil {
		return err
	}
	soaSerial = serial
	return nil
}

func DeleteRecord(db *sql.DB, id int) error {
	tx, err := db.Begin()
	_, err = tx.Exec("DELETE FROM dns_records WHERE id = ?", id)
	if err != nil {
		return err
	}
	return IncrementSerial(tx)
}

func UpdateRecord(db *sql.DB, id int, record dns.RR) error {
	tx, err := db.Begin()
	jsonString, err := json.Marshal(record)
	if err != nil {
		return err
	}
	_, err = tx.Exec("UPDATE dns_records SET name = ?, rrtype = ?, content = ? WHERE id = ?", record.Header().Name, record.Header().Rrtype, jsonString, id)
	if err != nil {
		return err
	}
	return IncrementSerial(tx)
}

func InsertRecord(db *sql.DB, record dns.RR) error {
	tx, err := db.Begin()
	jsonString, err := json.Marshal(record)
	if err != nil {
		return err
	}
	_, err = tx.Exec("INSERT INTO dns_records (name, rrtype, content) VALUES (?, ?, ?)", record.Header().Name, record.Header().Rrtype, jsonString)
	if err != nil {
		return err
	}
	return IncrementSerial(tx)
}

func GetRecordsForName(db *sql.DB, name string) (map[int]dns.RR, error) {
	fmt.Println(name)
	rows, err := db.Query("SELECT id, content FROM dns_records WHERE name LIKE ?", "%"+name)
	if err != nil {
		return nil, err
	}
	records := make(map[int]dns.RR)
	for rows.Next() {
		var content []byte
		var id int
		err = rows.Scan(&id, &content)
		if err != nil {
			return nil, err
		}
		record, err := ParseRecord(content)
		if err != nil {
			return nil, err
		}
		records[id] = record
	}
	return records, nil
}

func LogRequest(db *sql.DB, request *dns.Msg, response *dns.Msg, src_ip net.IP, src_host string) error {
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return err
	}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return err
	}
	name := request.Question[0].Name
	_, err = db.Exec("INSERT INTO dns_requests (name, request, response, src_ip, src_host) VALUES (?, ?, ?, ?, ?)", name, jsonRequest, jsonResponse, src_ip.String(), src_host)
	if err != nil {
		return err
	}
	StreamRequest(name, jsonRequest, jsonResponse, src_ip.String(), src_host)
	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func StreamRequest(name string, request []byte, response []byte, src_ip string, src_host string) error {
	// get base domain
	parts := strings.Split(name, ".")
	start := max(0, len(parts)-4)
	base := strings.Join(parts[start:], ".")
	x := map[string]interface{}{
		"created_at": time.Now().Unix(),
		"request":    string(request),
		"response":   string(response),
		"src_ip":     src_ip,
		"src_host":   src_host,
	}
	jsonString, err := json.Marshal(x)
	if err != nil {
		return err
	}
	WriteToStreams(base, jsonString)
	return nil
}

func DeleteRequestsForDomain(db *sql.DB, domain string) error {
	_, err := db.Exec("DELETE FROM dns_requests WHERE name LIKE ?", "%"+domain)
	if err != nil {
		return err
	}
	return nil
}

func GetRequests(db *sql.DB, domain string) ([]map[string]interface{}, error) {
	rows, err := db.Query("SELECT id, created_at, request, response, src_ip, src_host FROM dns_requests WHERE name LIKE ? ORDER BY created_at DESC", "%"+domain)
	if err != nil {
		return make([]map[string]interface{}, 0), err
	}
	requests := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int
		var created_at string
		var request []byte
		var response []byte
		var src_ip string
		var src_host string
		err = rows.Scan(&id, &created_at, &request, &response, &src_ip, &src_host)
		if err != nil {
			return make([]map[string]interface{}, 0), err
		}
		// parse created at to unix time
		created_time, err := time.Parse("2006-01-02 15:04:05", created_at)
		if err != nil {
			return make([]map[string]interface{}, 0), err
		}
		x := map[string]interface{}{
			"id":         id,
			"created_at": created_time.Unix(),
			"request":    string(request),
			"response":   string(response),
			"src_ip":     src_ip,
			"src_host":   src_host,
		}
		requests = append(requests, x)
	}
	return requests, nil
}

func GetRecords(db *sql.DB, name string, rrtype uint16) ([]dns.RR, error) {
	// return cname records if they exist
	rows, err := db.Query("SELECT content FROM dns_records WHERE name = ? AND (rrtype = ? OR rrtype = 5)", name, rrtype)
	if err != nil {
		return make([]dns.RR, 0), err
	}
	var records []dns.RR
	for rows.Next() {
		var content []byte
		err = rows.Scan(&content)
		if err != nil {
			return make([]dns.RR, 0), err
		}
		record, err := ParseRecord(content)
		if err != nil {
			return make([]dns.RR, 0), err
		}
		records = append(records, record)
	}
	return records, nil
}
