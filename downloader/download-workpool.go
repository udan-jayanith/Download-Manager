package main

import (
	"log"
	"os"
	"strconv"
)

type DownloadWorkPool struct {
	downloads chan DownloadItem
}

func NewDownloadWorkPool() DownloadWorkPool {
	maxDownloadInstances, err := strconv.Atoi(os.Getenv("maxDownloadInstances"))
	if err != nil {
		log.Println("maxDownloadInstances env file error. expected a integer.")
		log.Fatal(err)
	}
	downloads := make(chan DownloadItem, maxDownloadInstances)

	go func() {
		for downloadItem := range downloads {
			go downloadItem.download()
		}
	}()

	return DownloadWorkPool{
		downloads: downloads,
	}
}

func (dwp *DownloadWorkPool) Download(downloadItem DownloadItem) {
	go func() {
		dwp.downloads <- downloadItem
	}()
}
