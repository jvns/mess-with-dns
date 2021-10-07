package main

import (
	"database/sql"
	"encoding/json"

	_ "github.com/go-sql-driver/mysql"
	"github.com/miekg/dns"
)

// connect to planetscale
func connectDev() *sql.DB {
	// get connection string from environment
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/messwithdns")
	if err != nil {
		panic(err.Error())
	}
	return db
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

func GetRecordsForName(db *sql.DB, name string) []dns.RR {
	rows, err := db.Query("SELECT content FROM dns_records WHERE name = ?", name)
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
