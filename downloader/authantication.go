package main

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
)

func newToken() (token string) {
	for range 10 {
		token += string(rune('0') + (rand.Int32() % (rune('z') - rune('0'))))
	}
	return token
}

func getToken() (token string, err error) {
	err = Sqlite.Execute(func(db *sqlx.DB) error {
		row := db.QueryRow(`
			SELECT Token FROM token LIMIT 1;
		`)
		return row.Scan(&token)
	})
	return token, err
}

func saveToken(token string) error {
	return Sqlite.Execute(func(db *sqlx.DB) error {
		_, err := db.Exec(`
			INSERT INTO token (Token)
			VALUES (?);
		`, token)
		return err
	})
}

func HttpTokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:"+os.Getenv("port"))
	w.Header().Add("Content-Type", "application/json")
	fmt.Fprintf(w, `
		{
			"token": "%s"
		}
	`, token)
}

func RequireAuthenticationToken(w http.ResponseWriter, r *http.Request) (ok bool) {
	authToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if authToken == token {
		return true
	}

	cookieToken, err := r.Cookie("token")
	if err == http.ErrNoCookie {
		WriteError(w, err.Error())
		return false
	} else if err != nil {
		WriteError(w, err.Error())
		return false
	} else if cookieToken.Value != token {
		WriteError(w, "Invalid cookie"+" Received cookie == "+cookieToken.Value)
		return false
	}

	return true
}
