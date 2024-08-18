package users

import (
	"database/sql"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	_ "modernc.org/sqlite"
)

type UserService struct {
	db       *sql.DB
	hashKey  []byte
	blockKey []byte
}

//go:embed create.sql
var create_sql string

func Init(dbFile string, hashKey string, blockKey string) (*UserService, error) {
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	_, err = db.Exec(create_sql)
	if err != nil {
		return nil, err
	}

	decodedHash, err := base64.StdEncoding.DecodeString(hashKey)
	if err != nil {
		return nil, err
	}
	if len(decodedHash) != 32 {
		return nil, fmt.Errorf("HASH_KEY must be 32 bytes")
	}
	decodedBlock, err := base64.StdEncoding.DecodeString(blockKey)
	if err != nil {
		return nil, err
	}
	if len(decodedBlock) != 32 {
		return nil, fmt.Errorf("BLOCK_KEY must be 32 bytes")
	}
	return &UserService{db: db, hashKey: decodedHash, blockKey: decodedBlock}, nil
}

func getExistingSubdomains(db *sql.DB, word string) ([]string, error) {
	var subdomains []string
	rows, err := db.Query("SELECT name FROM subdomains WHERE name LIKE $1", word+"%")
	if err != nil {
		return subdomains, err
	}
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return subdomains, err
		}
		subdomains = append(subdomains, name)
	}
	return subdomains, nil
}

//go:embed words.json
var words_json []byte
var words_cache []string

func getWords() []string {
	if len(words_cache) == 0 {
		err := json.Unmarshal(words_json, &words_cache)
		if err != nil {
			panic(err)
		}
	}
	return words_cache
}

func randomWord() string {
	// get random integer
	words := getWords()
	randint := rand.Intn(len(words))
	return words[randint]
}

func smallestMissing(prefix string, existing []string) string {
	// omg it's smallest missing omg omg omg
	map_existing := make(map[string]bool)
	for _, subdomain := range existing {
		map_existing[subdomain] = true
	}
	for i := 5; ; i++ {
		subdomain := fmt.Sprintf("%s%d", prefix, i)
		if _, ok := map_existing[subdomain]; !ok {
			return subdomain
		}
	}
}

func insertAvailableSubdomain(db *sql.DB) (string, error) {
	prefix := randomWord()
	existing, err := getExistingSubdomains(db, prefix)
	if err != nil {
		return "", err
	}
	subdomain := smallestMissing(prefix, existing)
	_, err = db.Exec("INSERT INTO subdomains (name) VALUES ($1)", subdomain)
	if err != nil {
		return "", err
	}
	return subdomain, nil
}

func (u UserService) CreateAvailableSubdomain() (string, error) {
	var err error
	// try 3 times before giving up
	// this is a hack to make sure we don't randomly get a subdomain that's
	// already taken
	subdomain, err := insertAvailableSubdomain(u.db)
	if err == nil {
		return subdomain, nil
	}
	subdomain, err = insertAvailableSubdomain(u.db)
	if err == nil {
		return subdomain, nil
	}
	subdomain, err = insertAvailableSubdomain(u.db)
	if err == nil {
		return subdomain, nil
	}
	return "", err
}
