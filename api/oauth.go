package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/securecookie"
	"golang.org/x/oauth2"
)

type UserCookie struct {
	User string `json:"user"`
}

type GithubUser struct {
	Login string `json:"login"`
}

func getSecureCookie() *securecookie.SecureCookie {
	decoded, err := base64.StdEncoding.DecodeString(os.Getenv("HASH_KEY"))
	if err != nil {
		panic(err)
	}
	hashKey := []byte(decoded)
	if len(hashKey) != 32 {
		panic("HASH_KEY must be 32 bytes")
	}
	decoded, err = base64.StdEncoding.DecodeString(os.Getenv("BLOCK_KEY"))
	if err != nil {
		panic(err)
	}
	blockKey := []byte(decoded)
	if len(blockKey) != 32 {
		panic("BLOCK_KEY must be 32 bytes")
	}
	return securecookie.New(hashKey, blockKey)
}

func oauthCallback(w http.ResponseWriter, r *http.Request) {
	// get code from query
	code := r.URL.Query().Get("code")
	if code == "" {
		returnError(w, fmt.Errorf("Code not found"), http.StatusBadRequest)
		return
	}
	conf := oauthConfig()
	// exchange code for token
	token, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		returnError(w, err, http.StatusInternalServerError)
		return
	}
	// get user name
	client := conf.Client(oauth2.NoContext, token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		returnError(w, err, http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()
	var user GithubUser
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		returnError(w, err, http.StatusInternalServerError)
		return
	}
	// create a secure cookie
	setCookie(w, r, user.Login)

	// redirect to index
	http.Redirect(w, r, "/", http.StatusFound)

}

func setCookie(w http.ResponseWriter, r *http.Request, subdomain string) {
	sc := getSecureCookie()
	encoded, err := sc.Encode("session", UserCookie{
		User: subdomain,
	})
	if err != nil {
		returnError(w, err, http.StatusInternalServerError)
		return
	}
	// write secure cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
		// 2 weeks
		MaxAge:   24 * 60 * 60 * 14,
		SameSite: http.SameSiteStrictMode,
	})
	// set a regular username cookie for use in JS
	http.SetCookie(w, &http.Cookie{
		Name:  "username",
		Value: subdomain,
		Path:  "/",
		// 2 weeks
		MaxAge:   24 * 60 * 60 * 14,
		SameSite: http.SameSiteStrictMode,
	})
}

func ReadSessionUsername(r *http.Request) (string, error) {
	// in the test environment , don't check secure cookie
	if r.Host == "localhost:8080" {
		cookie, err := r.Cookie("username")
		if cookie == nil || err != nil {
			return "", fmt.Errorf("no username cookie")
		}
		return cookie.Value, err
	}
	sc := getSecureCookie()
	var user UserCookie
	// get session cookie
	cookie, err := r.Cookie("session")
	if err != nil {
		return "", err
	}
	err = sc.Decode("session", cookie.Value, &user)
	if err != nil {
		return "", err
	}
	fmt.Println("user:", user.User)
	return user.User, nil
}

func oauthConfig() *oauth2.Config {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		panic("GITHUB_CLIENT_ID or GITHUB_CLIENT_SECRET not found")
	}
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"user:name"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		},
	}
}

func githubOauth(w http.ResponseWriter, r *http.Request) {
	conf := oauthConfig()
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)
}
