package main

import (
	"encoding/json"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
)

type TokenStruct struct{}

func newToken() TokenStruct {
	token := TokenStruct{}
	err := Sqlite.Execute(func(db *sqlx.DB) error {
		_, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS token (Token TEXT NOT NULL);
		`)
		return err
	})
	if err != nil {
		log.Println("token table SQL execution error")
		log.Fatal(err)
	}

	var rowCount int
	err = Sqlite.Execute(func(db *sqlx.DB) error {
		row := db.QueryRow(`
			SELECT COUNT(Token) FROM token;
		`)
		return row.Scan(&rowCount)
	})
	if err != nil {
		log.Println("token table SQL execution error")
		log.Fatal(err)
	} else if rowCount <= 0 {
		token.saveToken(token.newToken())
	}
	return token
}

func (_ *TokenStruct) newToken() (token string) {
	for range 60 {
		token += string(rune('0') + (rand.Int32() % (rune('z') - rune('0'))))
	}
	return token
}

func (_ *TokenStruct) getToken() (token string, err error) {
	err = Sqlite.Execute(func(db *sqlx.DB) error {
		row := db.QueryRow(`
			SELECT Token FROM token LIMIT 1;
		`)
		return row.Scan(&token)
	})
	return token, err
}

func (_ *TokenStruct) saveToken(token string) error {
	return Sqlite.Execute(func(db *sqlx.DB) error {
		_, err := db.Exec(`
			INSERT INTO token (Token)
			VALUES (?);
		`, token)
		return err
	})
}

func (_ *TokenStruct) changeToken(token string) error {
	return Sqlite.Execute(func(db *sqlx.DB) error {
		_, err := db.Exec(`
			UPDATE token SET Token = ?;
		`, token)
		return err
	})
}

var (
	Token = newToken()
)

func getTokenFromReq(r *http.Request) string {
	return strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
}

func RequireAuthenticationToken(w http.ResponseWriter, r *http.Request) (ok bool) {
	token, err := Token.getToken()
	if err != nil {
		log.Println(err)
		return false
	}
	if getTokenFromReq(r) == token {
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

func HandleAuth(mux *http.ServeMux) {
	mux.HandleFunc("/token/get", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:"+os.Getenv("port"))
		w.Header().Add("Content-Type", "application/json")
		token, err := Token.getToken()
		if err != nil {
			WriteError(w, err.Error())
			return
		}

		res := map[string]string{
			"token": token,
		}
		json.NewEncoder(w).Encode(&res)
	})

	mux.HandleFunc(`/token/change`, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:"+os.Getenv("port"))
		w.Header().Add("Content-Type", "application/json")

		token := Token.newToken()
		err := Token.changeToken(token)
		if err != nil {
			WriteError(w, err.Error())
		}
	})

	mux.HandleFunc("/token/is-valid", func(w http.ResponseWriter, r *http.Request) {
		AllowCrossOrigin(w)
		reqToken := r.FormValue("token")
		token, err := Token.getToken()
		if err != nil {
			WriteError(w, err.Error())
			return
		}

		res := map[string]bool{
			"is-valid": reqToken == token,
		}
		json.NewEncoder(w).Encode(&res)
	})
}
