package db

import (
	"database/sql"
	"sync"

	_ "modernc.org/sqlite"
)

type LockedDB struct {
	db *sql.DB
	mu sync.Mutex
}

func (ldb *LockedDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	ldb.mu.Lock()
	defer ldb.mu.Unlock()
	return ldb.db.Query(query, args...)
}

func (ldb *LockedDB) Close() error {
	ldb.mu.Lock()
	defer ldb.mu.Unlock()
	return ldb.db.Close()
}

func (ldb *LockedDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	ldb.mu.Lock()
	defer ldb.mu.Unlock()
	return ldb.db.Exec(query, args...)
}

func Connect(filename string) (*LockedDB, error) {
	db, err := sql.Open("sqlite", filename)
	if err != nil {
		return nil, err
	}
	ldb := &LockedDB{db: db, mu: sync.Mutex{}}
	return ldb, nil
}
