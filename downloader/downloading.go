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
	defer di.Close()
	downloadReq, err := httpDownloadReq(di.URL, di.Headers, di.TempFilePath)
	if err != nil {
		di.Update(0, 0, 0, err)
		return
	}
	di.PartialContent = downloadReq.PartialContent
	di.ContentLength = int64(downloadReq.ContentLength)
	di.updateContentLength()

	download(downloadReq.Req, di.TempFilePath, di, di.cancel)
	if di.Status == Paused {
		return
	}

	saveFilepath := filepath.Join(di.Dir, di.FileName)
	err = os.Rename(di.TempFilePath, saveFilepath)
	if err != nil {
		di.Update(0, 0, 0, err)
		return
	}
	di.TempFilePath = ""
	di.updateTempFilepath()

	di.Update(0, 0, 0, nil)
}

type UpdateChan interface {
	Update(bps, length, estimatedTime int, err error)
	ChangeStatus(status DownloadStatus) error
}

func download(req *http.Request, destFilepath string, updates UpdateChan, cancelChan chan struct{}) {
	updates.ChangeStatus(Pending)

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

	destFileStat, err := destFile.Stat()
	if err != nil {
		updates.Update(0, 0, 0, err)
		return
	}
	destFile.Seek(destFileStat.Size(), 0)

	rd := bufio.NewReader(res.Body)
	t := time.Now()
	bps := 0
	length := int(destFileStat.Size())

	updates.ChangeStatus(Downloading)
	for {
		p := make([]byte, 1024)
		n1, err := rd.Read(p)
		if err != nil {
			break
		}

		select {
		case <-cancelChan:
			updates.ChangeStatus(Paused)
			updates.Update(0, 0, 0, nil)
			return
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
	updates.ChangeStatus(Complete)
}

func getHttpReqRes(req *http.Request) (*http.Response, error) {
	res, err := http.DefaultClient.Do(req)
	res.Body.Close()
	return res, err
}

type DownloadReq struct {
	Req            *http.Request
	ContentLength  int
	PartialContent bool
}

func httpDownloadReq(url string, headers []HTTPHeader, destFilepath string) (DownloadReq, error) {
	downloadReq := *new(DownloadReq)

	req, err := http.NewRequest("GET", url, nil)
	for _, header := range headers {
		req.Header.Set(header.Name, header.Value)
	}
	downloadReq.Req = req

	reqCopy := *req
	res, err := getHttpReqRes(&reqCopy)
	if err != nil {
		return downloadReq, err
	}

	//Set contentLength
	contentLength, err := strconv.Atoi(res.Header.Get("Content-Length"))
	downloadReq.ContentLength = contentLength

	//Set partialContent
	if res.Header.Get("Accept-Ranges") == "bytes" || res.StatusCode == http.StatusPartialContent {
		stat, err := os.Stat(destFilepath)
		if err != nil {
			return downloadReq, err
		}

		req.Header.Add("Range", fmt.Sprintf(`bytes=%v-%v`, stat.Size(), contentLength-1))
		downloadReq.PartialContent = true
	} else {
		err := os.Truncate(destFilepath, 0)
		if err != nil {
			return downloadReq, err
		}
	}
	return downloadReq, err
}
