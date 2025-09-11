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
* Websockets client.
* Resume, Pause, Get-downloading, get-downloads endpoints.
* Password protection.
* Improve UpdatesHandler marshaling.
*/

var (
	_                = godotenv.Load("./.env")
	downloadWorkPool = NewDownloadWorkPool()
	updatesHandler   = UpdatesHandler{
		maxConnections: 8,
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

	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			FileName string `json:"file-name"`
			URL      string `json:"url"`
			Dir      string `json:"dir"`
		}

		var req request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Fatal(err)
		}

		downloadItem := NewDownloadItem(req.FileName, req.Dir, req.URL)
		downloadWorkPool.Download(downloadItem)
	})


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
		updatesHandler.conns[conn] = struct{}{}
	})

	http.ListenAndServe(os.Getenv("port"), nil)
}
