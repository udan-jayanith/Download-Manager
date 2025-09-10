package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"bufio"
)

type DownloadStatus int

const (
	Pending DownloadStatus = iota
	Downloading
	Complete
)

type DownloadItem struct {
	ID       int64
	FileName string
	URL      string
	Dir      string
	//ContentLength length should be updated after the complete download.
	ContentLength int64
	DateAndTime   time.Time
	Status        DownloadStatus
}

func NewDownloadItem(FileName, Dir, URL string) DownloadItem {
	downloadItem := DownloadItem{
		FileName:    FileName,
		URL:         URL,
		Dir:         Dir,
		DateAndTime: time.Now(),
		Status:      Pending,
	}

	/*
		ID INTEGER PRIMARY KEY,
		FileName TEXT NOT NULL,
		URL TEXT NOT NULL,
		Dir TEXT NOT NULL,
		ContentLength INTEGER,
		DateAndTime TEXT NOT NULL
	*/
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
	return downloadItem
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

func (di *DownloadItem) download() {
	di.changeStatus(Downloading)
	defer di.changeStatus(Complete)

	res, err := http.Get(di.URL)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	rd := bufio.NewReader(res.Body) 
	file := di.newPackage()
	rd.WriteTo(file)
	defer file.Close()
}

//saveAs read packages in order and write to one file named FileName and store it in Dir by creating necessary parent dir.
// func (di *DownloadItem) saveAs()   {}
// func (di *DownloadItem) delete ()   {}

func (di *DownloadItem) newPackage() *os.File {
	downloadBufferDir := os.Getenv("downloadBuffDir")
	os.MkdirAll(downloadBufferDir, 0775)

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

	downloadItemFolder := filepath.Join(downloadBufferDir, strconv.Itoa(int(di.ID)))
	os.MkdirAll(downloadItemFolder, 0775)

	Packs++
	downloadItemPack := filepath.Join(downloadItemFolder, strconv.Itoa(int(Packs)))
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
	return file
}
