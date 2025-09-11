package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"

	"net/http"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

/*
# TODOS
* Resume, Pause, Get-downloading, get-downloads endpoints.
* Password protection.
* Improve UpdatesHandler marshaling.
*/

var (
	_                = godotenv.Load("./.env")
	downloadWorkPool = NewDownloadWorkPool()
	updatesHandler   = UpdatesHandler{
		maxConnections: 2,
		updatesChan:    downloadWorkPool.Updates,
		conns:          make(map[*websocket.Conn]struct{}, 1),
	}
)

func main() {
	Sqlite.Execute(func(db *sql.DB) {
		_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS downloads 
		(
		ID INTEGER PRIMARY KEY,
		FileName TEXT NOT NULL,
		URL TEXT NOT NULL,
		Dir TEXT NOT NULL,
		ContentLength INTEGER,
		DateAndTime TEXT NOT NULL,
		Packs INTEGER NOT NULL,
		Status INTEGER NOT NULL
		);`)
		if err != nil {
			log.Println("Creation of download table")
			log.Fatal(err)
		}
	})

	updatesHandler.Handle()

	waUpgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	http.HandleFunc("/wa/updates", func(w http.ResponseWriter, r *http.Request) {
		conn, err := waUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		err = updatesHandler.AddConn(conn)
		if err != nil {
			conn.Close()
		}
	})

	http.HandleFunc("/download", downloadHandler)
	http.HandleFunc("/get-downloads", downloadHandler)
	http.HandleFunc("/get-downloading", downloadHandler)
	http.HandleFunc("/resume", downloadHandler)
	http.HandleFunc("/pause", downloadHandler)
	http.HandleFunc("/remove", downloadHandler)

	http.ListenAndServe(os.Getenv("port"), nil)
}

func AllowCrossOrigin(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		FileName string `json:"file-name"`
		URL      string `json:"url"`
		Dir      string `json:"dir"`
		Password string `json:"password"`
	}

	var req request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Fatal(err)
	}

	if req.Password != os.Getenv("password") {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	AllowCrossOrigin(w)

	downloadItem := NewDownloadItem(req.FileName, req.Dir, req.URL)
	downloadWorkPool.Download(downloadItem)
}
