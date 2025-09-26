package main

type DownloadWorkPool struct {
	downloads     chan *DownloadItem
	downloadItems map[int64]*DownloadItem
	Updates       chan DownloadItemUpdate
}

func NewDownloadWorkPool() DownloadWorkPool {
	workPool := DownloadWorkPool{
		downloads:     make(chan *DownloadItem, 3),
		downloadItems: make(map[int64]*DownloadItem, 3),
		Updates:       make(chan DownloadItemUpdate, 8),
	}

	//Set a limit on how many concurrent downloads can run
	go func() {
		concurrentDownloadsChan := make(chan struct{}, 3)
		for downloadItem := range workPool.downloads {
			if downloadItem.deleted {
				delete(workPool.downloadItems, downloadItem.ID)
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
		dwp.downloadItems[downloadItem.ID] = &downloadItem
		dwp.downloads <- &downloadItem
	}()
}

func (dwp *DownloadWorkPool) GetDownloadItem(ID int64) (downloadItem *DownloadItem, ok bool) {
	downloadItem, ok = dwp.downloadItems[ID]
	return downloadItem, ok
}
