package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/miekg/dns"
)

// connect to planetscale
func connect() (*sql.DB, error) {
	// get connection string from environment
	connStr := os.Getenv("POSTGRES_CONNECTION_STRING")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	// important to avoid running out of memory
	db.SetMaxOpenConns(8)
	return db, nil
}

func createTables(db *sql.DB) error {
	if os.Getenv("DEV") == "true" {
		fmt.Println("creating tables...")
		err := loadSQLFile(db, "api/create.sql")
		if err != nil {
			return err
		}
		// initialize the serials table
		// check if serials table has anything in it
		rows, err := db.Query("SELECT * FROM dns_serials")
		defer rows.Close()
		if err != nil {
			return err
		}
		if rows.Next() {
			// if it has something in it, we don't need to do anything
			return nil
		}
		_, err = db.Exec("INSERT INTO dns_serials (serial) VALUES (10)")
		if err != nil {
			return err
		}
	}
	return nil
}

func loadSQLFile(db *sql.DB, sqlFile string) error {
	file, err := ioutil.ReadFile(sqlFile)
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		tx.Rollback()
	}()
	for _, q := range strings.Split(string(file), ";") {
		q := strings.TrimSpace(q)
		if q == "" {
			continue
		}
		if _, err := tx.Exec(q); err != nil {
			return err
		}
	}
	return tx.Commit()
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
	_, err = tx.Exec("DELETE FROM dns_records WHERE id = $1", id)
	if err != nil {
		return err
	}
	return IncrementSerial(tx)
}

func DeleteOldRecords(db *sql.DB) {
	// delete records where created_at timestamp is more than a week old
	_, err := db.Exec("DELETE FROM dns_records WHERE created_at < NOW() - '1 week'::interval")
	if err != nil {
		panic(err)
	}
}

func GetTotalRecords(db *sql.DB, parent string) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM dns_records WHERE name LIKE $1", "%"+parent).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func DeleteOldRequests(db *sql.DB) {
	// delete requests where created_at timestamp is more than a day
	// if we don't put the limit I get a "resources exhausted" error
	// 1 day ago, postgres
	_, err := db.Exec("DELETE FROM dns_requests WHERE created_at < NOW() - '1 week'::interval")
	if err != nil {
		panic(err)
	}
}

func UpdateRecord(db *sql.DB, id int, record dns.RR) error {
	tx, err := db.Begin()
	jsonString, err := json.Marshal(record)
	if err != nil {
		return err
	}
	name := record.Header().Name
	_, err = tx.Exec("UPDATE dns_records SET name = $1, subdomain = $2, rrtype = $3, content = $4 WHERE id = $5", strings.ToLower(name), ExtractSubdomain(name), record.Header().Rrtype, jsonString, id)
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
	name := record.Header().Name
	_, err = tx.Exec("INSERT INTO dns_records (name, subdomain, rrtype, content) VALUES ($1, $2, $3, $4)", strings.ToLower(name), ExtractSubdomain(name), record.Header().Rrtype, jsonString)
	if err != nil {
		return err
	}
	return IncrementSerial(tx)
}

func uncommittedTransation(db *sql.DB) (*sql.Tx, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec("SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED")
	if err != nil {
		return nil, err
	}
	return tx, nil
}

type Record struct {
	ID     int    `json:"id"`
	Record dns.RR `json:"record"`
}

func GetRecordsForName(db *sql.DB, subdomain string) ([]Record, error) {
	// we're stricter about the isolation level here because it's weird if you delete a record
	// but it still exists after
	rows, err := db.Query("SELECT id, content FROM dns_records WHERE subdomain = $1 ORDER BY created_at DESC", subdomain)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	records := make([]Record, 0)
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
		records = append(records, Record{ID: id, Record: record})
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
	subdomain := ExtractSubdomain(name)
	StreamRequest(subdomain, jsonRequest, jsonResponse, src_ip.String(), src_host)
	_, err = db.Exec("INSERT INTO dns_requests (name, subdomain, request, response, src_ip, src_host) VALUES ($1, $2, $3, $4, $5, $6)", name, subdomain, jsonRequest, jsonResponse, src_ip.String(), src_host)
	if err != nil {
		return err
	}
	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func StreamRequest(subdomain string, request []byte, response []byte, src_ip string, src_host string) error {
	// get base domain
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
	WriteToStreams(subdomain, jsonString)
	return nil
}

func DeleteRequestsForDomain(db *sql.DB, subdomain string) error {
	_, err := db.Exec("DELETE FROM dns_requests WHERE subdomain = $1", subdomain)
	if err != nil {
		return err
	}
	return nil
}

func GetRequests(db *sql.DB, subdomain string) ([]map[string]interface{}, error) {
	tx, err := uncommittedTransation(db)
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query("SELECT id, extract(epoch from created_at), request, response, src_ip, src_host FROM dns_requests WHERE subdomain = $1 ORDER BY created_at DESC LIMIT 30", subdomain)
	if err != nil {
		return make([]map[string]interface{}, 0), err
	}
	requests := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int
		var created_at float32
		var request []byte
		var response []byte
		var src_ip string
		var src_host string
		err = rows.Scan(&id, &created_at, &request, &response, &src_ip, &src_host)
		if err != nil {
			return make([]map[string]interface{}, 0), err
		}
		x := map[string]interface{}{
			"id":         id,
			"created_at": int64(created_at),
			"request":    string(request),
			"response":   string(response),
			"src_ip":     src_ip,
			"src_host":   src_host,
		}
		requests = append(requests, x)
	}
	tx.Commit()
	return requests, nil
}

func GetRecords(db *sql.DB, name string, rrtype uint16) ([]dns.RR, error) {
	tx, err := uncommittedTransation(db)
	if err != nil {
		return nil, err
	}
	// first get all the records
	rows, err := tx.Query("SELECT content FROM dns_records WHERE name = $1 ORDER BY created_at DESC", strings.ToLower(name))
	if err != nil {
		return nil, err
	}
	// next parse them
	var records []dns.RR
	for rows.Next() {
		var content []byte
		err = rows.Scan(&content)
		if err != nil {
			return nil, err
		}
		record, err := ParseRecord(content)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	// now filter them
	filtered := make([]dns.RR, 0)
	for _, record := range records {
		if shouldReturn(rrtype, record.Header().Rrtype) {
			filtered = append(filtered, record)
		}
	}
	tx.Commit()
	return filtered, nil
}

func shouldReturn(queryType uint16, recordType uint16) bool {
	if queryType == recordType {
		return true
	}
	if recordType == dns.TypeCNAME {
		return true
	}
	if queryType == dns.TypeHTTPS && (recordType == dns.TypeA || recordType == dns.TypeAAAA) {
		return true
	}
	return false
}
