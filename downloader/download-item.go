package main

import (
	"encoding/json"
	"log"
	"os"
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

func DownloadStatusToString(status DownloadStatus) string {
	switch status {
	case Pending:
		return "pending"
	case Downloading:
		return "downloading"
	case Complete:
		return "complete"
	case Paused:
		return "paused"
	}
	return ""
}

func DownloadStatusParse(status string) DownloadStatus {
	switch status {
	case "pending":
		return Pending
	case "downloading":
		return Downloading
	case "complete":
		return Complete
	case "paused":
		return Paused
	}
	return -1
}

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
	updateJson.Status = DownloadStatusToString(diu.Status)

	return json.Marshal(&updateJson)
}

type DownloadItem struct {
	ID             int64
	FileName       string
	URL            string
	Dir            string
	PartialContent bool
	ContentLength  int64
	DateAndTime    time.Time
	Status         DownloadStatus

	Updates      chan DownloadItemUpdate
	TempFilePath string
	cancel       chan struct{}
	deleted      bool
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

		cancel:  make(chan struct{}, 1),
		deleted: false,
	}

	tempFile, err := os.CreateTemp(os.TempDir(), "Download-Manager")
	tempFile.Close()
	downloadItem.TempFilePath = tempFile.Name()
	if err != nil {
		downloadItem.Update(0, 0, 0, err)
		return downloadItem
	}
	downloadItem.updateTempFilepath()

	err = Sqlite.Execute(func(db *sqlx.DB) error {
		results, err := db.Exec(`
			INSERT INTO downloads (FileName, URL, Dir, ContentLength, DateAndTime, Status, TempFilePath)
			VALUES (?, ?, ?, 0, ?, ?, ?);
		`, downloadItem.FileName, downloadItem.URL, downloadItem.Dir, downloadItem.DateAndTime.Format(time.RFC3339), Pending, tempFile.Name())
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
	downloadItem.Status = Pending
	downloadItem.updateStatus()
	downloadItem.Update(0, 0, 0, nil)
	return downloadItem
}

func (di *DownloadItem) updateStatus() error {
	return Sqlite.Execute(func(db *sqlx.DB) error {
		_, err := db.Exec(`
			UPDATE downloads SET Status = ? WHERE ID = ?;
		`, di.Status, di.ID)
		return err
	})
}

func (di *DownloadItem) ChangeStatus(status DownloadStatus) error {
	di.Status = status
	return di.updateStatus()
}

func (di *DownloadItem) updateContentLength() error {
	return Sqlite.Execute(func(db *sqlx.DB) error {
		_, err := db.Exec(`
			UPDATE downloads SET ContentLength = ? WHERE ID = ?;
		`, di.ContentLength, di.ID)
		return err
	})
}

func (di *DownloadItem) updateTempFilepath() error {
	return Sqlite.Execute(func(db *sqlx.DB) error {
		_, err := db.Exec(`
			UPDATE downloads SET TempFilePath = ? WHERE ID = ?;
		`, di.TempFilePath, di.ID)
		return err
	})
}

func (di *DownloadItem) Cancel() {
	di.deleted = true
	di.cancel <- struct{}{}
}

// Delete delete the downloadItem from the database. This also delete the save files and any files created.
func (di *DownloadItem) Delete() {
	os.Remove(di.TempFilePath)

	Sqlite.Execute(func(db *sqlx.DB) error {
		_, err := db.Exec(`
			DELETE FROM downloads WHERE ID = ?;
		`, di.ID)
		return err
	})
}

func (di *DownloadItem) Close() {
	close(di.cancel)
	close(di.Updates)
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

func (dij *DownloadItemJson) ToDownloadItem() (downloadItem DownloadItem, err error) {
	downloadItem = DownloadItem{
		ID:             dij.ID,
		FileName:       dij.FileName,
		URL:            dij.URL,
		Dir:            dij.Dir,
		PartialContent: dij.PartialContent,
		ContentLength:  dij.ContentLength,
		TempFilePath:   dij.TempFilePath,

		deleted: false,
		Updates: make(chan DownloadItemUpdate, 8),
		cancel:  make(chan struct{}, 1),
	}
	//Date and time and Status
	t, err := time.Parse(time.RFC3339, dij.DateAndTime)
	if err != nil {
		return downloadItem, err
	}
	downloadItem.DateAndTime = t
	downloadItem.Status = DownloadStatusParse(dij.Status)

	return downloadItem, nil
}

func (di *DownloadItem) JSON() DownloadItemJson {
	var v DownloadItemJson
	v.ID = di.ID
	v.FileName = di.FileName
	v.URL = di.URL
	v.Dir = di.Dir
	v.ContentLength = di.ContentLength
	v.DateAndTime = di.DateAndTime.Format(time.RFC3339)
	v.Status = DownloadStatusToString(di.Status)
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
