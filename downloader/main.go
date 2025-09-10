package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"

	"net/http"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

var (
	downloadWorkPool = NewDownloadWorkPool()
	_ = godotenv.Load("./.env")
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

	http.ListenAndServe(os.Getenv("port"), nil)
}
