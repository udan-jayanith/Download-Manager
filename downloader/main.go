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
	http.HandleFunc("/get-downloading", getDownloading)
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
	w.Header().Add("Content-Type", "application/json")
	fmt.Fprintf(w, `
		{
			"error": %s
		}
	`, errorMsg)
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

func RowToDownloadItem(rows *sql.Rows) (DownloadItem, error) {
	var downloadItem DownloadItem
	var dateAndTime string
	rows.Scan(&downloadItem.ID, &downloadItem.FileName, &downloadItem.URL, &downloadItem.Dir, &downloadItem.ContentLength, &dateAndTime, &downloadItem.Status)
	if dateAndTime != "" {
		t, err := time.Parse(time.RFC3339, dateAndTime)
		if err != nil {
			return downloadItem, err
		}
		downloadItem.DateAndTime = t
	}
	return downloadItem, nil
}

func getDownloads(w http.ResponseWriter, r *http.Request) {
	AllowCrossOrigin(w)

	limit := 20
	dateAndTime := r.FormValue("date-and-time")

	Sqlite.Mutex.Lock()
	defer Sqlite.Mutex.Unlock()

	var sqliteRows *sql.Rows
	if strings.TrimSpace(dateAndTime) == "" {
		rows, err := Sqlite.DB.Query(fmt.Sprintf(`
				SELECT * FROM downloads WHERE Status = %v ORDER BY DateAndTime ASC LIMIT %v;
			`, Complete, limit))
		if err != nil {
			WriteError(w, err.Error())
			return
		}
		sqliteRows = rows
	} else {
		rows, err := Sqlite.DB.Query(fmt.Sprintf(`
				SELECT * FROM downloads WHERE Status = %v AND DateAndTime > datetime('%s') ORDER BY DateAndTime ASC LIMIT %v;
			`, Complete, dateAndTime, limit))
		if err != nil {
			WriteError(w, err.Error())
			return
		}
		sqliteRows = rows
	}
	defer sqliteRows.Close()

	type JsonResponse struct {
		DownloadItems []DownloadItemJson `json:"download-items"`
	}
	jsonRes := JsonResponse{
		DownloadItems: make([]DownloadItemJson, 0, 20),
	}

	for sqliteRows.Next() {
		downloadItem, err := RowToDownloadItem(sqliteRows)
		if err != nil {
			WriteError(w, err.Error())
			return
		}

		jsonRes.DownloadItems = append(jsonRes.DownloadItems, downloadItem.JSON())
	}
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&jsonRes)
}

func getDownloading(w http.ResponseWriter, r *http.Request) {
	AllowCrossOrigin(w)

	type JsonRes struct {
		DownloadingItems []DownloadItemJson `json:"downloading-items"`
	}
	jsonRes := JsonRes{
		DownloadingItems: make([]DownloadItemJson, 0, 3),
	}

	err := Sqlite.Execute(func(db *sql.DB) error {
		rows, err := db.Query(`
			SELECT * FROM downloads WHERE Status = ?;
		`, Downloading)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			downloadItem, err := RowToDownloadItem(rows)
			if err != nil {
				return err
			}
			jsonRes.DownloadingItems = append(jsonRes.DownloadingItems, downloadItem.JSON())
		}
		return err
	})
	if err != nil {
		WriteError(w, err.Error())
		return
	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&jsonRes)
}
