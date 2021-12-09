package main

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
)

func getExistingSubdomains(db *sql.DB, word string) ([]string, error) {
	var subdomains []string
	rows, err := db.Query("SELECT name FROM subdomains WHERE name LIKE ?", word+"%")
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

func getShortWords(words []string) []string {
	shortwords := make([]string, 0)
	for _, word := range words {
		if len(word) <= 7 {
			shortwords = append(shortwords, word)
		}
	}
	return shortwords
}

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
	_, err = db.Exec("INSERT INTO subdomains (name) VALUES (?)", subdomain)
	if err != nil {
		return "", err
	}
	return subdomain, nil
}

func createAvailableSubdomain(db *sql.DB) (string, error) {
	var err error
	// try 3 times before giving up
	// this is a hack to make sure we don't randomly get a subdomain that's
	// already taken
	subdomain, err := insertAvailableSubdomain(db)
	if err == nil {
		return subdomain, nil
	}
	subdomain, err = insertAvailableSubdomain(db)
	if err == nil {
		return subdomain, nil
	}
	subdomain, err = insertAvailableSubdomain(db)
	if err == nil {
		return subdomain, nil
	}
	return "", err
}

func loginRandom(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	subdomain, err := createAvailableSubdomain(db)

	if err != nil {
		returnError(w, err, http.StatusInternalServerError)
		return
	}
	setCookie(w, r, subdomain)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
