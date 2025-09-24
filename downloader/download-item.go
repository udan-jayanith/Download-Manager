package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
)

type DownloadStatus int

const (
	Pending DownloadStatus = iota
	Downloading
	Complete
	Paused
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
	case Paused:
		updateJson.Status = "paused"
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
	TempFilePath   string
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

	tempFile, err := os.CreateTemp(os.TempDir(), "Download-Manager")
	if err != nil {
		downloadItem.Update(0, 0, 0, err)
		return downloadItem
	}
	defer tempFile.Close()

	tempFilePath := filepath.Join(os.TempDir(), tempFile.Name())
	err = Sqlite.Execute(func(db *sqlx.DB) error {
		results, err := db.Exec(`
			INSERT INTO downloads (FileName, URL, Dir, ContentLength, DateAndTime, Status, TempFilePath)
			VALUES (?, ?, ?, 0, ?, ?, ?);
		`, downloadItem.FileName, downloadItem.URL, downloadItem.Dir, downloadItem.DateAndTime.Format(time.RFC3339), Pending, tempFilePath)
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
		downloadItem.Update(0, 0, 0, err)
		return downloadItem
	}
	downloadItem.Update(0, 0, 0, nil)
	return downloadItem
}

func (di *DownloadItem) changeStatus(status DownloadStatus) error {
	di.Status = status
	return Sqlite.Execute(func(db *sqlx.DB) error {
		_, err := db.Exec(fmt.Sprintf(`
			UPDATE downloads SET Status = %v WHERE ID = %v	
		`, int(status), di.ID))
		return err
	})
}

type DownloadItemJson struct {
	ID             int64  `json:"id" db:"ID"`
	FileName       string `json:"file-name" db:"FileName"`
	URL            string `json:"url" db:"URL"`
	Dir            string `json:"dir" db:"Dir"`
	ContentLength  int64  `json:"content-length" db:"ContentLength"`
	DateAndTime    string `json:"date-and-time" db:"DateAndTime"`
	Status         string `json:"status" db:"Status"`
	PartialContent bool   `json:"partial-content" db:"FileName"`
	TempFilePath   string `db:"TempFilePath"`
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

func (di *DownloadItem) Update(bps, length, estimatedTime int, err error) {
	di.Updates <- DownloadItemUpdate{
		DownloadID:     di.ID,
		ContentLength:  int(di.ContentLength),
		Status:         di.Status,
		PartialContent: di.PartialContent,

		BytesPerSec:   bps,
		Length:        length,
		EstimatedTime: estimatedTime,
		Err:           err,
	}
}
