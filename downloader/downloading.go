package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func (di *DownloadItem) download() {
	di.Status = Downloading
	di.updateStatus()

	tempFile, err := os.OpenFile(di.TempFilePath, os.O_RDWR, 0777)
	if err != nil {
		di.Update(0, 0, 0, err)
		di.Status = Complete
		di.updateStatus()
		return
	}
	defer tempFile.Close()

	req, err := http.NewRequest("GET", di.URL, nil)
	req.Header.Add("Range", "bytes=0")

	cancelChan := make(chan struct{})
	download(req, tempFile, di, cancelChan)
	close(cancelChan)
	tempFile.Close()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		di.Update(0, 0, 0, err)
		return
	}

	saveFilepath := filepath.Join(homeDir, di.Dir, di.FileName)
	err = os.Rename(di.TempFilePath, saveFilepath)
	if err != nil {
		di.Update(0, 0, 0, err)
		return
	}

	di.TempFilePath = ""
	di.updateTempFilepath()

	di.updateContentLength()
	di.Status = Complete

	di.Update(0, 0, 0, nil)
}

type UpdateChan interface {
	Update(bps, length, estimatedTime int, err error)
	setContentLength(contentLength int64)
}

func download(req *http.Request, destFile *os.File, updates UpdateChan, cancelChan chan struct{}) {
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		updates.Update(0, 0, 0, err)
		return 
	}
	defer res.Body.Close()

	contentLength, err := strconv.Atoi(res.Header.Get("Content-Length"))
	updates.setContentLength(int64(contentLength))
	if err != nil {
		updates.Update(0, 0, 0, err)
		return 
	}

	rd := bufio.NewReader(res.Body)
	t := time.Now()
	bps := 0
	length := 0

loop:
	for {
		p := make([]byte, 1024)
		n1, err := rd.Read(p)
		if err != nil {
			break
		}

		select {
		case <-cancelChan:
			break loop
		default:
		}

		n2, err := destFile.Write(p[:n1])
		if err != nil {
			updates.Update(0, 0, 0, err)
			break
		} else if n1 != n2 {
			updates.Update(0, 0, 0, fmt.Errorf("Writing and reading miss matched."))
			break
		}

		length += n1
		bps += n1

		if time.Since(t).Seconds() >= 1 {
			updates.Update(bps, length, func() int {
				if bps <= 0 {
					return 0
				}
				return (contentLength - length) / bps
			}(), nil)

			bps = 0
			t = time.Now()
		}
	}
}
