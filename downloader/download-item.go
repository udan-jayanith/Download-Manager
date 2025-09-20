package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type DownloadStatus int

const (
	Pending DownloadStatus = iota
	Downloading
	Complete
)

type DownloadItemUpdate struct {
	DownloadID     int64          `json:"download-id"` //required
	BytesPerSec    int            `json:"bps"`
	ContentLength  int            `json:"content-length"`
	Length         int            `json:"length"`
	EstimatedTime  int            `json:"estimated-time"`
	Status         DownloadStatus `json:"download-status"`
	PartialContent bool           `json:"partial-content"`
	Err            error          `json:"error"`
}

func (diu *DownloadItemUpdate) JSON() ([]byte, error) {
	type UpdateJson struct {
		DownloadID     int    `json:"download-id"`
		Bps            int    `json:"bps"`
		ContentLength  int    `json:"content-length"`
		Length         int    `json:"length"`
		EstimatedTime  int    `json:"estimated-time"`
		Status         string `json:"status"`
		PartialContent bool   `json:"partial-content"`
		Err            string `json:"error"`
	}

	updateJson := UpdateJson{}
	updateJson.DownloadID = int(diu.DownloadID)
	updateJson.Bps = diu.BytesPerSec
	updateJson.ContentLength = diu.ContentLength
	updateJson.Length = diu.Length
	updateJson.EstimatedTime = diu.EstimatedTime
	updateJson.PartialContent = diu.PartialContent
	if diu.Err != nil {
		updateJson.Err = diu.Err.Error()
	} else {
		updateJson.Err = ""
	}

	switch diu.Status {
	case Pending:
		updateJson.Status = "pending"
	case Downloading:
		updateJson.Status = "downloading"
	case Complete:
		updateJson.Status = "complete"
	}

	return json.Marshal(&updateJson)
}

type DownloadItem struct {
	ID       int64
	FileName string
	URL      string
	Dir      string
	//ContentLength length should be updated after the complete download.
	ContentLength  int64
	DateAndTime    time.Time
	Status         DownloadStatus
	Updates        chan DownloadItemUpdate
	PartialContent bool
}

func NewDownloadItem(FileName, Dir, URL string) DownloadItem {
	downloadItem := DownloadItem{
		FileName:       FileName,
		URL:            URL,
		Dir:            Dir,
		DateAndTime:    time.Now(),
		Status:         Pending,
		Updates:        make(chan DownloadItemUpdate, 8),
		PartialContent: false,
	}

	err := Sqlite.Execute(func(db *sql.DB) error {
		results, err := db.Exec(`
			INSERT INTO downloads (FileName, URL, Dir, ContentLength, DateAndTime, Status)
			VALUES (?, ?, ?, 0, ?, ?);
		`, downloadItem.FileName, downloadItem.URL, downloadItem.Dir, downloadItem.DateAndTime.Format(time.RFC3339), Pending)
		if err != nil {
			return err
		}

		ID, err := results.LastInsertId()
		if err != nil {
			log.Println("NewDownloadItem function")
			return err
		}
		downloadItem.ID = ID
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	downloadItem.Updates <- DownloadItemUpdate{
		DownloadID: downloadItem.ID,
		Status:     Pending,
	}
	return downloadItem
}

// tempDir returns the temDir location.
func (di *DownloadItem) tempDir() string {
	tempDir := filepath.Join(os.TempDir(), "DownloadManager", strconv.Itoa(int(di.ID)))
	os.MkdirAll(tempDir, 0o700)
	return tempDir
}

func (di *DownloadItem) changeStatus(status DownloadStatus) error {
	di.Status = status
	return Sqlite.Execute(func(db *sql.DB) error {
		_, err := db.Exec(fmt.Sprintf(`
			UPDATE downloads SET Status = %v WHERE ID = %v	
		`, int(status), di.ID))
		if err != nil {
			log.Println("Downloading Status updating error.")
		}
		return err
	})
}

type DownloadItemJson struct {
	ID             int64  `json:"id"`
	FileName       string `json:"file-name"`
	URL            string `json:"url"`
	Dir            string `json:"dir"`
	ContentLength  int64  `json:"content-length"`
	DateAndTime    string `json:"date-and-time"`
	Status         string `json:"status"`
	PartialContent bool   `json:"partial-content"`
}

func (di *DownloadItem) JSON() DownloadItemJson {
	var v DownloadItemJson
	v.ID = di.ID
	v.FileName = di.FileName
	v.URL = di.URL
	v.Dir = di.Dir
	v.ContentLength = di.ContentLength
	v.DateAndTime = di.DateAndTime.Format(time.RFC3339)
	switch di.Status {
	case Pending:
		v.Status = "pending"
	case Downloading:
		v.Status = "downloading"
	case Complete:
		v.Status = "complete"
	}
	v.PartialContent = di.PartialContent

	return v
}
