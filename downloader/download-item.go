package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	DownloadID    int64          `json:"download-id"` //required
	BytesPerSec   int            `json:"bps"`
	ContentLength int            `json:"content-length"`
	Length        int            `json:"length"`
	EstimatedTime time.Duration  `json:"estimated-time"`
	Status        DownloadStatus `json:"download-status"`
}

func (diu *DownloadItemUpdate) JSON() ([]byte, error) {
	type UpdateJson struct {
		DownloadID    int     `json:"download-id"`
		Bps           int     `json:"bps"`
		ContentLength int     `json:"content-length"`
		Length        int     `json:"length"`
		EstimatedTime float64 `json:"estimated-time"`
		Status        string  `json:"status"`
	}

	updateJson := UpdateJson{}
	updateJson.DownloadID = int(diu.DownloadID)
	updateJson.Bps = diu.BytesPerSec
	updateJson.ContentLength = diu.ContentLength
	updateJson.Length = diu.Length
	updateJson.EstimatedTime = diu.EstimatedTime.Seconds()

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
	ContentLength int64
	DateAndTime   time.Time
	Status        DownloadStatus
	Updates       chan DownloadItemUpdate
}

func NewDownloadItem(FileName, Dir, URL string) DownloadItem {
	downloadItem := DownloadItem{
		FileName:    FileName,
		URL:         URL,
		Dir:         Dir,
		DateAndTime: time.Now(),
		Status:      Pending,
		Updates:     make(chan DownloadItemUpdate, 8),
	}

	Sqlite.Execute(func(db *sql.DB) {
		results, err := db.Exec(fmt.Sprintf(`
			INSERT INTO downloads (FileName, URL, Dir, DateAndTime, Packs, Status)
			VALUES ('%s', '%s', '%s', '%s', 0, 0);
		`, downloadItem.FileName, downloadItem.URL, downloadItem.Dir, downloadItem.DateAndTime.String()))
		if err != nil {
			log.Fatal(err)
		}

		ID, err := results.LastInsertId()
		if err != nil {
			log.Panicln("NewDownloadItem function")
			log.Fatal(err)
		}
		downloadItem.ID = ID
	})
	downloadItem.Updates <- DownloadItemUpdate{
		DownloadID: downloadItem.ID,
		Status:     Pending,
	}
	return downloadItem
}

func (di *DownloadItem) download() {
	di.changeStatus(Downloading)
	di.Updates <- DownloadItemUpdate{
		DownloadID: di.ID,
		Status:     Downloading,
	}
	defer func() {
		di.Updates <- DownloadItemUpdate{
			DownloadID: di.ID,
			Status:     Complete,
		}
		di.changeStatus(Complete)
	}()

	res, err := http.Get(di.URL)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	//TODO
	rd := bufio.NewReader(res.Body)
	file, tempDir := di.newPackage()
	rd.WriteTo(file)
	defer file.Close()

	di.save(tempDir)
	file.Close()
	err = os.RemoveAll(tempDir)
	if err != nil {
		log.Println("Temp dir deletion error.")
		log.Fatal(err)
	}
}

func (di *DownloadItem) save(tempDir string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Println("Home dir error")
		log.Fatal(err)
	}

	fileDir := filepath.Join(homeDir, di.Dir)
	os.MkdirAll(fileDir, 0775)
	os.Remove(filepath.Join(fileDir, di.FileName))

	saveFilePath := filepath.Join(fileDir, di.FileName+".DownloadManager")
	os.Remove(saveFilePath)
	saveFile, err := os.Create(saveFilePath)
	if err != nil {
		log.Println("File creation error")
		log.Fatal(err)
	}
	defer saveFile.Close()

	var Packs int64
	Sqlite.Execute(func(db *sql.DB) {
		row := db.QueryRow(fmt.Sprintf(`SELECT Packs FROM downloads WHERE ID = %v LIMIT 1;`, di.ID))
		err := row.Scan(&Packs)
		if err != nil {
			log.Fatal(err)
		}
	})
	for i := int64(1); i <= Packs; i++ {
		downloadPackage, err := os.Open(filepath.Join(tempDir, strconv.Itoa(int(i))))
		if err != nil {
			log.Fatal(err)
		}
		downloadPackage.WriteTo(saveFile)
		downloadPackage.Close()
	}
	saveFile.Close()
	os.Rename(saveFilePath, filepath.Join(fileDir, di.FileName))
}

func (di *DownloadItem) newPackage() (*os.File, string) {
	tempDir := filepath.Join(os.TempDir(), "DownloadManager")
	os.Mkdir(tempDir, 0o700)
	tempDir, err := os.MkdirTemp(tempDir, strconv.Itoa(int(di.ID)))
	if err != nil {
		log.Println("Temp dir error.")
		log.Fatal(err)
	}

	var Packs int64
	Sqlite.Execute(func(db *sql.DB) {
		row := db.QueryRow(fmt.Sprintf(`
			SELECT Packs FROM downloads WHERE ID = %v;
		`, di.ID))
		err := row.Scan(&Packs)
		if err != nil {
			log.Println("newPackage function")
			log.Fatal(err)
		}
	})

	Packs++
	downloadItemPack := filepath.Join(tempDir, strconv.Itoa(int(Packs)))
	os.Remove(downloadItemPack)

	Sqlite.Execute(func(db *sql.DB) {
		_, err := db.Exec(fmt.Sprintf(`
			UPDATE downloads SET Packs = %v WHERE ID = %v;
		`, Packs, di.ID))
		if err != nil {
			log.Println("Updating packs count error")
			log.Fatal(err)
		}
	})

	file, err := os.Create(downloadItemPack)
	if err != nil {
		log.Println("newPackage function")
		log.Fatal(err)
	}
	return file, tempDir
}

func (di *DownloadItem) changeStatus(status DownloadStatus) {
	Sqlite.Execute(func(db *sql.DB) {
		_, err := db.Exec(fmt.Sprintf(`
			UPDATE downloads SET Status = %v WHERE ID = %v	
		`, int(status), di.ID))
		if err != nil {
			log.Fatal(err)
		}
	})
}