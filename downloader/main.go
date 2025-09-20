package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"net/http"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

/*
# TODOS
* Close channels and update the db when a download is completed with the content length ot the package.
* Get-downloading, get-downloads endpoints.
* Password protection.
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
	err := Sqlite.Execute(func(db *sql.DB) error {
		_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS downloads 
		(
		ID INTEGER PRIMARY KEY,
		FileName TEXT NOT NULL,
		URL TEXT NOT NULL,
		Dir TEXT NOT NULL,
		ContentLength INTEGER,
		DateAndTime TEXT NOT NULL,
		Status INTEGER NOT NULL
		);`)
		return err
	})
	if err != nil {
		log.Println("downloads table SQL execution error")
		log.Fatal(err)
	}

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
	http.HandleFunc("/get-downloads", getDownloads)
	http.HandleFunc("/get-downloading", downloadHandler)
	http.HandleFunc("/resume", downloadHandler)
	http.HandleFunc("/pause-resume", downloadHandler)

	http.ListenAndServe(os.Getenv("port"), nil)
}

func AllowCrossOrigin(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func WriteError(w http.ResponseWriter, errorMsg string) {
	w.Write([]byte(fmt.Sprintf(`
		{
			"error": %s
		}
	`, errorMsg)))
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
		WriteError(w, fmt.Sprintf("body decoding error, %s", err))
		return
	}

	if req.Password != os.Getenv("password") {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	AllowCrossOrigin(w)

	downloadItem := NewDownloadItem(req.FileName, req.Dir, req.URL)
	downloadWorkPool.Download(downloadItem)
}

func getDownloads(w http.ResponseWriter, r *http.Request) {
	AllowCrossOrigin(w)

	limit := 20
	dateAndTime := r.FormValue("date-and-time")

	if strings.TrimSpace(dateAndTime) == "" {
		err := Sqlite.Execute(func(db *sql.DB) error {
			rows, err := db.Query(fmt.Sprintf(`
				SELECT * FROM downloads ORDER BY DateAndTime ASC LIMIT %v;
			`, limit))
			if err != nil {
				return err
			}
			defer rows.Close()

			for rows.Next() {
				var downloadItem DownloadItem
				var dateAndTime string
				rows.Scan(&downloadItem.ID, &downloadItem.FileName, &downloadItem.URL, &downloadItem.Dir, &downloadItem.ContentLength, &dateAndTime, &downloadItem.Status)
				log.Println(downloadItem, dateAndTime)
			}
			return err
		})
		if err != nil {
			WriteError(w, "Querying error.")
			log.Println(err)
			return
		}

		return
	} else {
		err := Sqlite.Execute(func(db *sql.DB) error {
			rows, err := db.Query(fmt.Sprintf(`
				SELECT * FROM downloads WHERE DateAndTime > datetime('%s') ORDER BY DateAndTime ASC LIMIT %v;
			`, dateAndTime, limit))
			if err != nil {
				return err
			}
			defer rows.Close()

			/*
				ID INTEGER PRIMARY KEY,
				FileName TEXT NOT NULL,
				URL TEXT NOT NULL,
				Dir TEXT NOT NULL,
				ContentLength INTEGER,
				DateAndTime TEXT NOT NULL,
				Status INTEGER NOT NULL
			*/
			for rows.Next() {
				var downloadItem DownloadItem
				var dateAndTime string
				rows.Scan(&downloadItem.ID, &downloadItem.FileName, &downloadItem.URL, &downloadItem.Dir, &downloadItem.ContentLength, &dateAndTime, &downloadItem.Status)
				t, err := time.Parse(time.RFC3339, dateAndTime)
				if err != nil && dateAndTime != "" {
					return err
				}
				downloadItem.DateAndTime = t
				log.Println(downloadItem)
			}
			return err
		})
		if err != nil {
			WriteError(w, "Querying error.")
			log.Println(err)
			return
		}
	}

}
