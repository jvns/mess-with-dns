package main

import (
	"encoding/base64"
	"fmt"

	"github.com/gorilla/securecookie"
)

func main() {
	key := securecookie.GenerateRandomKey(32)
	fmt.Println(base64.StdEncoding.EncodeToString(key))
	key = securecookie.GenerateRandomKey(32)
	fmt.Println(base64.StdEncoding.EncodeToString(key))
}
