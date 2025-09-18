package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// newFile creates a new file. If the file already exists newFile delete it and create a new file in the path named by fileName.
func newFile(filePath string) (*os.File, error) {
	os.MkdirAll(filepath.Dir(filePath), 0664)
	os.RemoveAll(filePath)

	return os.Create(filePath)
}

func (di *DownloadItem) download() {
	header, err := getHeaders(di.URL)
	if err != nil {
		log.Println("Error while getting the header.")
		log.Println(err)
		return
	}

	contentLength, err := strconv.Atoi(header.Get("Content-Length"))
	if err != nil {
		contentLength = 0
	}

	if header.Get("Content-Length") == "" || header.Get("Accept-Ranges") == "none" || header.Get("Accept-Ranges") == "" || contentLength <= 0 {
		err = di.sequentialDownload(contentLength)
		if err != nil {
			log.Println("Error occurred while downloading sequentially")
			log.Println(err)
		}
		return
	}
	err = di.parallelDownload(contentLength)
	if err != nil {
		log.Println("parallelDownload error")
		log.Println(err)
	}
}

func getHeaders(url string) (http.Header, error) {
	res, err := http.Get(url)
	res.Body.Close()
	return res.Header, err
}

func (di *DownloadItem) sequentialDownload(contentLength int) error {
	err := di.changeStatus(Downloading)
	if err != nil {
		log.Println("sequentialDownload function")
		return err
	}
	di.PartialContent = false
	di.Updates <- DownloadItemUpdate{
		DownloadID:    di.ID,
		Status:        Downloading,
		ContentLength: contentLength,
	}
	defer func() error {
		di.Updates <- DownloadItemUpdate{
			DownloadID:    di.ID,
			Status:        Complete,
			ContentLength: contentLength,
			Length:        contentLength,
		}
		err := di.changeStatus(Complete)
		return err
	}()

	res, err := http.Get(di.URL)
	if err != nil {
		log.Println(err)
		return err
	}
	defer res.Body.Close()

	downloadItemUpdate := DownloadItemUpdate{
		DownloadID:    di.ID,
		ContentLength: contentLength,
		Status:        Downloading,
	}

	rd := bufio.NewReader(res.Body)
	filePath := filepath.Join(di.tempDir(), "1")
	file, err := newFile(filePath)
	if err != nil {
		log.Println(err)
		return err
	}

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
			downloadItemUpdate.EstimatedTime = func() int {
				if bps > 0 && contentLength-length > 0 {
					return (contentLength - length) / bps
				}
				return 0
			}()
			di.Updates <- downloadItemUpdate

			t = time.Now()
			bps = 0
		}
		_, err = file.Write([]byte{byt})
		if err != nil {
			log.Println("File writing error.")
			return err
		}
	}
	file.Close()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Println("Home dir error")
		return err
	}
	err = os.Rename(filePath, filepath.Join(homeDir, di.Dir, di.FileName))
	err = os.RemoveAll(di.tempDir())

	if err != nil {
		log.Println("Temp dir deletion error.")
	}
	return err
}

func (di *DownloadItem) parallelDownload(contentLength int) error {
	err := di.changeStatus(Downloading)
	if err != nil {
		return err
	}
	di.PartialContent = true
	di.Updates <- DownloadItemUpdate{
		DownloadID:     di.ID,
		Status:         di.Status,
		ContentLength:  contentLength,
		PartialContent: true,
	}
	defer func() {
		di.Updates <- DownloadItemUpdate{
			DownloadID:     di.ID,
			Status:         Complete,
			ContentLength:  contentLength,
			Length:         contentLength,
			PartialContent: true,
		}
		di.changeStatus(Complete)
	}()

	chunks := 4
	chunkSize := (contentLength - 1) / chunks
	//This return both bytes form both position and in between.
	offsetLeft := 0
	offsetRight := chunkSize - 1
	if offsetRight <= chunks || offsetRight <= 0 {
		return di.sequentialDownload(contentLength)
	}

	wg := &sync.WaitGroup{}
	length := atomic.Int64{}
	bps := atomic.Int64{}

	wg.Go(func() {
		for int(length.Load()) < contentLength {
			time.Sleep(time.Second)

			downloadItemUpdate := DownloadItemUpdate{
				DownloadID:    di.ID,
				BytesPerSec:   int(bps.Load()),
				ContentLength: contentLength,
				Length:        int(length.Load()),
				EstimatedTime: func() int {
					if bps.Load() > 0 && contentLength-int(length.Load()) > 0 {
						return int((int64(contentLength) - length.Load())/bps.Load())
					}
					return 0
				}(),
				Status:         di.Status,
				PartialContent: true,
			}
			di.Updates <- downloadItemUpdate
			bps.Store(0)
		}
	})

	fileNo := 0
	for offsetLeft < contentLength {
		fileNo++
		startPos, endPos := offsetLeft, offsetRight
		currentFileNo := fileNo
		wg.Go(func() {
			//Http request
			req, err := http.NewRequest("GET", di.URL, nil)
			if err != nil {
				log.Println("Parallel download request error")
				log.Fatal(err)
			}
			req.Header.Add("Range", fmt.Sprintf("bytes=%v-%v", startPos, endPos))

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Println(err)
				return
			} else if res.StatusCode != http.StatusPartialContent {
				log.Println("Got unexpected status code expected", http.StatusPartialContent, "but got", res.StatusCode)
			}
			defer res.Body.Close()

			rd := bufio.NewReader(res.Body)
			file, err := newFile(filepath.Join(di.tempDir(), strconv.Itoa(currentFileNo)))
			if err != nil {
				log.Println("Error occurred when calling newFile.")
				log.Fatal(err)
			}
			defer file.Close()

			for {
				byt, err := rd.ReadByte()
				if err != nil {
					break
				}

				length.Add(1)
				bps.Add(1)
				_, err = file.Write([]byte{byt})
				if err != nil {
					log.Println("Writing error")
					log.Println(err)
					return
				}
			}
		})
		offsetLeft += chunkSize
		offsetRight = min(offsetLeft+chunkSize-1, contentLength-1)
	}

	wg.Wait()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Println("Home dir error")
		return err
	}
	saveFilePath := filepath.Join(homeDir, di.Dir, di.FileName+".download-manager")
	os.MkdirAll(filepath.Dir(saveFilePath), 0775)
	os.RemoveAll(saveFilePath)
	saveFile, err := os.Create(saveFilePath)
	if err != nil {
		log.Println("Error occurred while creating a save file.")
		return err
	}
	defer saveFile.Close()

	for i := 1; i <= fileNo; i++ {
		file, err := os.Open(filepath.Join(di.tempDir(), strconv.Itoa(i)))
		if err != nil {
			log.Println("Error occurred while opening a file.")
			file.Close()
			return err
		}

		_, err = file.WriteTo(saveFile)
		if err != nil {
			log.Println("Writing error")
			return err
		}
		file.Close()
	}
	saveFile.Close()

	newSaveFilePath := filepath.Join(homeDir, di.Dir, di.FileName)
	os.RemoveAll(newSaveFilePath)
	err = os.Rename(saveFilePath, newSaveFilePath)
	if err != nil {
		log.Println("Renaming error error")
		return err
	}

	err = os.RemoveAll(di.tempDir())
	if err != nil {
		log.Println("TempDir removing error")
	}
	return err
}
