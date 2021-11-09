package main

import (
	"database/sql"
	"encoding/json"
    "fmt"
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

func DeleteRecord(db *sql.DB, id int) {
	_, err := db.Exec("DELETE FROM dns_records WHERE id = ?", id)
	if err != nil {
		panic(err.Error())
	}
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
}

func GetRecordsForName(db *sql.DB, name string) map[int]dns.RR {
    fmt.Println(name)
	rows, err := db.Query("SELECT id, content FROM dns_records WHERE name = ?", name)
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

func GetRecords(db *sql.DB, name string, rrtype uint16) []dns.RR {
	rows, err := db.Query("SELECT content FROM dns_records WHERE name = ? AND rrtype = ?", name, rrtype)
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
