package main

import (
	"log"
	"os"
	"strconv"
)

type DownloadWorkPool struct {
	downloads chan DownloadItem
	Updates   chan DownloadItemUpdate
}

func NewDownloadWorkPool() DownloadWorkPool {
	maxDownloadInstances, err := strconv.Atoi(os.Getenv("maxDownloadInstances"))
	if err != nil {
		log.Println("maxDownloadInstances env file error. expected a integer.")
		log.Fatal(err)
	}
	workPool := DownloadWorkPool{
		downloads: make(chan DownloadItem, maxDownloadInstances),
		Updates:   make(chan DownloadItemUpdate, 8),
	}

	//Set a limit on how many concurrent downloads can run
	go func() {
		concurrentDownloadsChan := make(chan struct{}, 3)
		for downloadItem := range workPool.downloads {
			concurrentDownloadsChan <- struct{}{}
			go func() {
				for update := range downloadItem.Updates {
					workPool.Updates <- update
					if update.Status == Complete {
						break
					}
				}
			}()
			go func() {
				downloadItem.download()
				<-concurrentDownloadsChan
			}()
		}
		close(workPool.downloads)
		close(workPool.Updates)
	}()

	return workPool
}

func (dwp *DownloadWorkPool) Download(downloadItem DownloadItem) {
	go func() {
		dwp.downloads <- downloadItem
	}()
}
