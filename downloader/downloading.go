package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func (di *DownloadItem) download() {
	header, err := getHeaders(di.URL)
	if err != nil {
		log.Println(err)
		return
	}

	contentLength, err := strconv.Atoi(header.Get("Content-Length"))
	if err != nil {
		contentLength = 0
	}
	//if header.Get("Content-Length") == "" || header.Get("Accept-Ranges") == "none" || header.Get("Accept-Ranges") == "" {
	di.sequentialDownload(contentLength)
	//return
	//}

}

func getHeaders(url string) (http.Header, error) {
	res, err := http.Get(url)
	res.Body.Close()
	return res.Header, err
}

func (di *DownloadItem) sequentialDownload(contentLength int) {
	di.changeStatus(Downloading)
	di.Updates <- DownloadItemUpdate{
		DownloadID: di.ID,
		Status:     Downloading,
		ContentLength: contentLength,
	}
	defer func() {
		di.Updates <- DownloadItemUpdate{
			DownloadID: di.ID,
			Status:     Complete,
			ContentLength: contentLength,
		}
		di.changeStatus(Complete)
	}()

	res, err := http.Get(di.URL)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	downloadItemUpdate := DownloadItemUpdate{
		DownloadID:    di.ID,
		ContentLength: contentLength,
		Status:        Downloading,
	}

	rd := bufio.NewReader(res.Body)
	file, tempDir := di.newPackage()

	bps := 0
	length := 0
	t := time.Now()
	for {
		byt, err := rd.ReadByte()
		if err != nil {
			break
		}
		length++
		bps++
		if time.Since(t).Seconds() >= 1 {
			downloadItemUpdate.Length = length
			downloadItemUpdate.BytesPerSec = bps
			downloadItemUpdate.EstimatedTime = (contentLength - length) / bps
			di.Updates <- downloadItemUpdate

			t = time.Now()
			bps = 0
		}
		file.Write([]byte{byt})
	}
	defer file.Close()

	di.save(tempDir)
	file.Close()
	err = os.RemoveAll(tempDir)
	if err != nil {
		log.Println("Temp dir deletion error.")
		log.Fatal(err)
	}
}

func (di *DownloadItem) parallelDownload() {
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

}
