package main

import "sync"

type DownloadWorkPool struct {
	downloads     chan *DownloadItem
	downloadItems map[int64]*DownloadItem
	Updates       chan DownloadItemUpdate
	mutex         *sync.Mutex
}

func NewDownloadWorkPool() DownloadWorkPool {
	workPool := DownloadWorkPool{
		downloads:     make(chan *DownloadItem, 3),
		downloadItems: make(map[int64]*DownloadItem, 3),
		Updates:       make(chan DownloadItemUpdate, 8),
		mutex:         &sync.Mutex{},
	}

	//Set a limit on how many concurrent downloads can run
	go func() {
		concurrentDownloadsChan := make(chan struct{}, 3)
		for downloadItem := range workPool.downloads {
			if downloadItem.deleted {
				workPool.mutex.Lock()
				delete(workPool.downloadItems, downloadItem.ID)
				workPool.mutex.Unlock()
				continue
			}

			concurrentDownloadsChan <- struct{}{}
			//Handle updates to the download item
			go func() {
				for update := range downloadItem.Updates {
					workPool.Updates <- update
					if update.Status == Complete || update.Status == Paused {
						break
					}
				}
			}()
			//Download the download content
			go func() {
				downloadItem.download()
				workPool.mutex.Lock()
				defer workPool.mutex.Unlock()
				delete(workPool.downloadItems, downloadItem.ID)
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
		dwp.mutex.Lock()
		defer dwp.mutex.Unlock()
		dwp.downloadItems[downloadItem.ID] = &downloadItem
		dwp.downloads <- &downloadItem
	}()
}

func (dwp *DownloadWorkPool) GetDownloadItem(ID int64) (downloadItem *DownloadItem, ok bool) {
	dwp.mutex.Lock()
	defer dwp.mutex.Unlock()
	downloadItem, ok = dwp.downloadItems[ID]
	return downloadItem, ok
}
