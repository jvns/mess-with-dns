package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
    "net"
    "time"
	"os"
    "strings"

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
    StreamRequest(name, jsonRequest, jsonResponse, src_ip.String(), src_host)
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}

func StreamRequest(name string, request []byte, response []byte, src_ip string, src_host string) {
    // get base domain
    parts := strings.Split(name, ".")
    start := max(0, len(parts) - 4)
    base := strings.Join(parts[start:], ".")
    x := map[string]interface{}{
        "created_at": time.Now().Unix(),
        "request": string(request),
        "response": string(response),
        "src_ip": src_ip,
        "src_host": src_host,
    }
    jsonString, err := json.Marshal(x)
    if err != nil {
        panic(err.Error())
        return
    }
    WriteToStreams(base, jsonString)
}

func GetRequests(db *sql.DB, domain string) []map[string]interface{} {
    rows, err := db.Query("SELECT id, created_at, request, response, src_ip, src_host FROM dns_requests WHERE name LIKE ?", "%" + domain)
    if err != nil {
        panic(err.Error())
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
        // parse created at to unix time
        created_time, err := time.Parse("2006-01-02 15:04:05", created_at)
        if err != nil {
            panic(err.Error())
        }
        if err != nil {
            panic(err.Error())
        }
        x := map[string]interface{}{
            "id": id,
            "created_at": created_time.Unix(),
            "request": string(request),
            "response": string(response),
            "src_ip": src_ip,
            "src_host": src_host,
        }
        requests = append(requests, x)
    }
    return requests
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
