package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"net/http"

	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func HandleDownloads(mux *http.ServeMux) {
	err := Sqlite.Execute(func(db *sqlx.DB) error {
		_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS downloads 
		(
		ID INTEGER PRIMARY KEY,
		FileName TEXT NOT NULL,
		URL TEXT NOT NULL,
		Dir TEXT NOT NULL,
		ContentLength INTEGER,
		DateAndTime TEXT NOT NULL,
		Status INTEGER NOT NULL,
		TempFilePath TEXT
		);`)
		return err
	})
	if err != nil {
		log.Println("downloads table SQL execution error")
		log.Fatal(err)
	}

	err = Sqlite.Execute(func(db *sqlx.DB) error {
		_, err := db.Exec(`
			UPDATE downloads SET Status = ?
			WHERE Status = ? OR Status = ?;	
		`, Paused, Downloading, Pending)
		return err
	})
	if err != nil {
		log.Println("downloads table corruptions fix error")
		log.Fatal(err)
	}

	waUpgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	mux.HandleFunc("/download/wa/updates", func(w http.ResponseWriter, r *http.Request) {
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

	mux.HandleFunc("/download/download", downloadHandler)
	mux.HandleFunc("/download/get-downloads", getDownloads)
	mux.HandleFunc("/download/get-downloading", getDownloading)
	mux.HandleFunc("/download/search-downloads", searchDownload)
	mux.HandleFunc("/download/get-download-item", getDownloadItem)
	mux.HandleFunc("/download/pause", pauseDownload)
	mux.HandleFunc("/download/resume", resumeDownload)
	mux.HandleFunc("/download/delete", deleteDownload)
}

type HTTPHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type DownloadRequest struct {
	FileName string       `json:"file-name"`
	URL      string       `json:"url"`
	Dir      string       `json:"dir"`
	Headers  []HTTPHeader `json:"headers"`
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	AllowCrossOrigin(w)
	if !RequireAuthenticationToken(w, r) {
		return
	}

	var req DownloadRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		WriteError(w, fmt.Sprintf("body decoding error, %s", err))
		return
	} else if req.Headers == nil {
		req.Headers = make([]HTTPHeader, 0)
	}

	if strings.TrimSpace(req.FileName) == "" || strings.TrimSpace(req.Dir) == "" || strings.TrimSpace(req.URL) == "" {
		WriteError(w, "Missing file-name, dir or url.")
		return
	}

	req.Dir, err = Dir(req.Dir)
	if err != nil {
		WriteError(w, err.Error())
		return
	}
	downloadItem := NewDownloadItem(req)
	downloadWorkPool.Download(downloadItem)
}

func getDownloads(w http.ResponseWriter, r *http.Request) {
	AllowCrossOrigin(w)
	if !RequireAuthenticationToken(w, r) {
		return
	}

	limit := 20
	dateAndTime := r.FormValue("date-and-time")

	Sqlite.Mutex.Lock()
	defer Sqlite.Mutex.Unlock()

	jsonRes := struct {
		DownloadItems []DownloadItemJson `json:"download-items"`
	}{
		DownloadItems: make([]DownloadItemJson, 0, 20),
	}

	if strings.TrimSpace(dateAndTime) == "" {
		err := Sqlite.DB.Select(&jsonRes.DownloadItems, `
				SELECT * FROM downloads WHERE Status = ? ORDER BY DateAndTime DESC LIMIT ?;
			`, Complete, limit)
		if err != nil {
			WriteError(w, err.Error())
			return
		}
	} else {
		err := Sqlite.DB.Select(&jsonRes.DownloadItems, `
				SELECT * FROM downloads WHERE Status = ? AND DateAndTime < ? ORDER BY DateAndTime DESC LIMIT ?;
			`, Complete, dateAndTime, limit)
		if err != nil {
			WriteError(w, err.Error())
			return
		}
	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&jsonRes)
}

func getDownloading(w http.ResponseWriter, r *http.Request) {
	AllowCrossOrigin(w)
	if !RequireAuthenticationToken(w, r) {
		return
	}

	jsonRes := struct {
		DownloadingItems []DownloadItemJson `json:"downloading-items"`
	}{
		DownloadingItems: make([]DownloadItemJson, 0, 3),
	}

	err := Sqlite.Execute(func(db *sqlx.DB) error {
		err := db.Select(&jsonRes.DownloadingItems, `
			SELECT * FROM downloads WHERE Status != ?;
		`, Complete)
		if err != nil {
			return err
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

func searchDownload(w http.ResponseWriter, r *http.Request) {
	AllowCrossOrigin(w)
	if !RequireAuthenticationToken(w, r) {
		return
	}

	query := strings.TrimSpace(r.FormValue("query"))
	if query == "" {
		WriteError(w, "Missing query")
		return
	}
	query = "%" + query + "%"

	Sqlite.Mutex.Lock()
	defer Sqlite.Mutex.Unlock()

	searchResults := struct {
		SearchResults []DownloadItemJson `json:"search-results"`
	}{
		SearchResults: make([]DownloadItemJson, 0, 20),
	}

	err := Sqlite.DB.Select(&searchResults.SearchResults, `
		SELECT * FROM downloads WHERE Status = ? AND (FileName LIKE ? OR URL LIKE ?) LIMIT 20;
	`, Complete, query, query)
	if err != nil {
		WriteError(w, err.Error())
		return
	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&searchResults)
}

func getDownloadItem(w http.ResponseWriter, r *http.Request) {
	AllowCrossOrigin(w)
	if !RequireAuthenticationToken(w, r) {
		return
	}

	downloadID, err := strconv.Atoi(r.FormValue("download-id"))
	if err != nil {
		WriteError(w, err.Error())
		return
	}

	downloadItem, ok := downloadWorkPool.GetDownloadItem(int64(downloadID))
	var downloadItemJson DownloadItemJson
	if !ok {
		err := Sqlite.Execute(func(db *sqlx.DB) error {
			return db.Get(&downloadItemJson, `
				SELECT * FROM downloads WHERE ID = ? LIMIT 1;
			`, downloadID)
		})
		if err != nil {
			WriteError(w, err.Error())
			return
		}
	} else {
		downloadItemJson = downloadItem.JSON()
	}
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&downloadItemJson)
}

func pauseDownload(w http.ResponseWriter, r *http.Request) {
	AllowCrossOrigin(w)
	if !RequireAuthenticationToken(w, r) {
		return
	}

	downloadID, err := strconv.Atoi(r.FormValue("download-id"))
	if err != nil {
		WriteError(w, err.Error())
		return
	}

	downloadItem, ok := downloadWorkPool.GetDownloadItem(int64(downloadID))
	if !ok {
		return
	} else if !downloadItem.PartialContent {
		WriteError(w, "Pause is not supported.")
		return
	}
	downloadItem.Cancel()
}

func resumeDownload(w http.ResponseWriter, r *http.Request) {
	AllowCrossOrigin(w)
	if !RequireAuthenticationToken(w, r) {
		return
	}

	downloadID, err := strconv.Atoi(r.FormValue("download-id"))
	if err != nil {
		WriteError(w, "Missing download-id")
		return
	}

	var downloadItemJson DownloadItemJson
	err = Sqlite.Execute(func(db *sqlx.DB) error {
		return db.Get(&downloadItemJson, `
				SELECT * FROM downloads WHERE ID = ? LIMIT 1;
			`, downloadID)
	})
	if err != nil {
		WriteError(w, err.Error())
		return
	} else if downloadItemJson.URL == "" {
		WriteError(w, "URL is not found.")
		return
	}

	downloadItem, err := downloadItemJson.ToDownloadItem()
	if err != nil {
		WriteError(w, err.Error())
		return
	}

	downloadWorkPool.Download(downloadItem)
}

func deleteDownload(w http.ResponseWriter, r *http.Request) {
	AllowCrossOrigin(w)
	if !RequireAuthenticationToken(w, r) {
		return
	}

	downloadId, err := strconv.Atoi(r.FormValue("download-id"))
	if err != nil {
		WriteError(w, err.Error())
		return
	}

	downloadItem, ok := downloadWorkPool.GetDownloadItem(int64(downloadId))
	if !ok {
		WriteError(w, "Download Item not found")
		return
	}
	downloadItem.Update(0, 0, 0, fmt.Errorf("deleted"))

	downloadItem.Cancel()
	downloadItem.Delete()
}
