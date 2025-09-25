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

	downloadReq, err := httpDownloadReq(di.URL, di.TempFilePath)
	if err != nil {
		di.Update(0, 0, 0, err)
		return
	}
	di.PartialContent = downloadReq.PartialContent
	di.ContentLength = int64(downloadReq.ContentLength)
	di.updateContentLength()

	cancelChan := make(chan struct{})
	download(downloadReq.Req, di.TempFilePath, di, cancelChan)
	close(cancelChan)
	di.updateContentLength()

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

	di.Status = Complete
	di.updateStatus()

	di.Update(0, 0, 0, nil)
}

type UpdateChan interface {
	Update(bps, length, estimatedTime int, err error)
}

func download(req *http.Request, destFilepath string, updates UpdateChan, cancelChan chan struct{}) {
	destFile, err := os.OpenFile(destFilepath, os.O_RDWR, 0777)
	if err != nil {
		updates.Update(0, 0, 0, err)
		return
	}
	defer destFile.Close()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		updates.Update(0, 0, 0, err)
		return
	}
	defer res.Body.Close()

	contentLength, err := strconv.Atoi(res.Header.Get("Content-Length"))
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

func getHttpReqHeader(url string) (*http.Response, error) {
	res, err := http.Get(url)
	res.Body.Close()
	return res, err
}

type DownloadReq struct {
	Req            *http.Request
	ContentLength  int
	PartialContent bool
}

func httpDownloadReq(url string, destFilepath string) (DownloadReq, error) {
	downloadReq := *new(DownloadReq)

	res, err := getHttpReqHeader(url)
	if err != nil {
		return downloadReq, err
	}

	//Set contentLength
	contentLength, err := strconv.Atoi(res.Header.Get("Content-Length"))
	downloadReq.ContentLength = contentLength
	if err != nil {
		return downloadReq, err
	}

	req, err := http.NewRequest("GET", url, nil)
	downloadReq.Req = req

	//Set partialContent
	if res.Header.Get("") == "bytes" {
		stat, err := os.Stat(destFilepath)
		if err != nil {
			return downloadReq, err
		}

		startingPosition := stat.Size()
		if startingPosition > 0 {
			startingPosition++
		}

		req.Header.Add("Range", fmt.Sprintf(`bytes=%v-%v`, startingPosition, contentLength-1))
		downloadReq.PartialContent = true
	} else {
		err := os.Truncate(destFilepath, 0)
		if err != nil {
			return downloadReq, err
		}
	}
	return downloadReq, err
}
