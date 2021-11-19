package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
    "net"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/miekg/dns"
)

// connect to planetscale
func connect() *sql.DB {
	// get connection string from environment
	db, err := sql.Open("mysql", os.Getenv("PLANETSCALE_CONNECTION_STRING"))
	if err != nil {
		panic(err.Error())
	}
	return db
}

func GetSerial(db *sql.DB) uint32 {
    var serial uint32
    err := db.QueryRow("SELECT serial FROM dns_serials").Scan(&serial)
    if err != nil {
        panic(err.Error())
    }
    return serial
}

func IncrementSerial(db *sql.DB) {
    _, err := db.Exec("UPDATE dns_serials SET serial = serial + 1")
    if err != nil {
        panic(err.Error())
    }
}

func DeleteRecord(db *sql.DB, id int) {
	_, err := db.Exec("DELETE FROM dns_records WHERE id = ?", id)
	if err != nil {
		panic(err.Error())
	}
    IncrementSerial(db)
}

func UpdateRecord(db *sql.DB, id int, record dns.RR) {
	jsonString, err := json.Marshal(record)
	if err != nil {
		panic(err.Error())
	}
	_, err = db.Exec("UPDATE dns_records SET name = ?, rrtype = ?, content = ? WHERE id = ?", record.Header().Name, record.Header().Rrtype, jsonString, id)
	if err != nil {
		panic(err.Error())
	}
    IncrementSerial(db)
}

func InsertRecord(db *sql.DB, record dns.RR) {
	jsonString, err := json.Marshal(record)
	if err != nil {
		panic(err.Error())
	}
	_, err = db.Exec("INSERT INTO dns_records (name, rrtype, content) VALUES (?, ?, ?)", record.Header().Name, record.Header().Rrtype, jsonString)
	if err != nil {
		panic(err.Error())
	}
    IncrementSerial(db)
}

func GetRecordsForName(db *sql.DB, name string) map[int]dns.RR {
	fmt.Println(name)
	rows, err := db.Query("SELECT id, content FROM dns_records WHERE name LIKE ?", "%" + name)
	if err != nil {
		panic(err.Error())
	}
	records := make(map[int]dns.RR)
	for rows.Next() {
		var content []byte
		var id int
		err = rows.Scan(&id, &content)
		if err != nil {
			panic(err.Error())
		}
		record, err := ParseRecord(content)
		if err != nil {
			panic(err.Error())
		}
		records[id] = record
	}
	return records
}

func LogRequest(db *sql.DB, request *dns.Msg, response *dns.Msg, src_ip net.IP, src_host string) {
    jsonRequest, err := json.Marshal(request)
    if err != nil {
        fmt.Println("error logging request: ", err.Error())
        return
    }
    jsonResponse, err := json.Marshal(response)
    if err != nil {
        fmt.Println("error logging request: ", err.Error())
        return
    }
    name := request.Question[0].Name
    _, err = db.Exec("INSERT INTO dns_requests (name, request, response, src_ip, src_host) VALUES (?, ?, ?, ?, ?)", name, jsonRequest, jsonResponse, src_ip.String(), src_host)
    if err != nil {
        fmt.Println("error logging request: ", err.Error())
    }
}

func GetRecords(db *sql.DB, name string, rrtype uint16) []dns.RR {
	// return cname records if they exist
	rows, err := db.Query("SELECT content FROM dns_records WHERE name = ? AND (rrtype = ? OR rrtype = 5)", name, rrtype)
	if err != nil {
		panic(err.Error())
	}
	var records []dns.RR
	for rows.Next() {
		var content []byte
		err = rows.Scan(&content)
		if err != nil {
			panic(err.Error())
		}
		record, err := ParseRecord(content)
		if err != nil {
			panic(err.Error())
		}
		records = append(records, record)
	}
	return records
}
