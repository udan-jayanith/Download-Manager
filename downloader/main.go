package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"net/http"

	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

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
	http.HandleFunc("/search-downloads", searchDownload)

	http.HandleFunc("/pause", func(w http.ResponseWriter, r *http.Request) {
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
	})

	http.HandleFunc("/resume", func(w http.ResponseWriter, r *http.Request) {
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
	})

	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		downloadId, err := strconv.Atoi(r.FormValue("download-id"))
		if err != nil {
			WriteError(w, err.Error())
			return
		}

		var downloadItemJson DownloadItemJson
		err = Sqlite.Execute(func(db *sqlx.DB) error {
			return db.Get(&downloadItemJson, `
				SELECT * FROM downloads WHERE ID = ? LIMIT 1;
			`, downloadId)
		})
		if err != nil {
			WriteError(w, err.Error())
			return
		}

		downloadItem, err := downloadItemJson.ToDownloadItem()
		if err != nil {
			WriteError(w, err.Error())
			return
		}

		downloadItem.Delete()
	})

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
			"error": "%s"
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

func getDownloads(w http.ResponseWriter, r *http.Request) {
	AllowCrossOrigin(w)

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
				SELECT * FROM downloads WHERE Status = ? ORDER BY DateAndTime ASC LIMIT ?;
			`, Complete, limit)
		if err != nil {
			WriteError(w, err.Error())
			return
		}
	} else {
		err := Sqlite.DB.Select(&jsonRes.DownloadItems, `
				SELECT * FROM downloads WHERE Status = ? AND DateAndTime > ? ORDER BY DateAndTime ASC LIMIT ?;
			`, Complete, dateAndTime, limit)
		if err != nil {
			WriteError(w, err.Error())
			return
		}
		log.Println(dateAndTime)
	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&jsonRes)
}

func getDownloading(w http.ResponseWriter, r *http.Request) {
	AllowCrossOrigin(w)

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
		SELECT * FROM downloads WHERE Status = ? AND (FileName LIKE ? OR URL LIKE ?);
	`, Complete, query, query)
	if err != nil {
		WriteError(w, err.Error())
		return
	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&searchResults)
}

//func HandleCorruptions()
