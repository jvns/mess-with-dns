package users

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/securecookie"
)

type UserCookie struct {
	User string `json:"user"`
}

func (u UserService) getSecureCookie() *securecookie.SecureCookie {
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
	return securecookie.New(u.hashKey, u.blockKey)
}

func (u UserService) SetCookie(w http.ResponseWriter, r *http.Request, subdomain string) {
	sc := u.getSecureCookie()
	encoded, err := sc.Encode("session", UserCookie{
		User: subdomain,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// write secure cookie
	http.SetCookie(w, &http.Cookie{
		Name:  "session",
		Value: encoded,
		Path:  "/",
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

func (u UserService) ReadSessionUsername(r *http.Request) (string, error) {
	sc := u.getSecureCookie()
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
