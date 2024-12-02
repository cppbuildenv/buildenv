package io

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
)

func NewDownloadRequest(url, destDir string) *downloadRequest {
	return &downloadRequest{
		url:     url,
		destDir: destDir,
	}
}

type downloadRequest struct {
	url         string
	destDir     string
	archiveName string
	offline     bool
}

func (d *downloadRequest) SetArchiveName(archiveName string) *downloadRequest {
	d.archiveName = archiveName
	return d
}

func (d *downloadRequest) SetOffline(offline bool) *downloadRequest {
	d.offline = offline
	return d
}

func (d downloadRequest) Download() (downloadedFile string, err error) {
	fileName, err := getFileName(d.url)
	if err != nil {
		return "", err
	}

	// In offline mode, it'll return the file path directly.
	if d.offline {
		downloadedFile = filepath.Join(d.destDir, fileName)
		if d.archiveName != "" && d.archiveName != fileName {
			downloadedFile = filepath.Join(d.destDir, d.archiveName)
		}
		return
	}

	// Read file size.
	resp, err := http.Get(d.url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	fileSize := resp.ContentLength
	progress := NewProgressBar(fileName, fileSize)

	if err := os.MkdirAll(d.destDir, os.ModeDir|os.ModePerm); err != nil {
		return "", err
	}

	// Build download file path.
	downloadedFile = filepath.Join(d.destDir, fileName)
	file, err := os.Create(downloadedFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Copy to local file with progress.
	_, err = io.Copy(io.MultiWriter(file, progress), resp.Body)
	if err != nil {
		return "", err
	}

	// Rename downloaded file if specified and not same as downloaded file.
	if d.archiveName != "" && d.archiveName != fileName {
		renameFile := filepath.Join(d.destDir, d.archiveName)
		if err := os.Rename(downloadedFile, renameFile); err != nil {
			return "", err
		}
		downloadedFile = renameFile
	}

	return downloadedFile, nil
}

func getFileName(downloadURL string) (string, error) {
	// Read file name from URL.
	u, err := url.Parse(downloadURL)
	if err != nil {
		return "", err
	}
	filename := path.Base(u.Path)
	if filename != "." && filename != "/" {
		return filename, nil
	}

	// Read file name from http header.
	resp, err := http.Head(downloadURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	re := regexp.MustCompile(`filename=["]?([^"]+)["]?`)
	header := resp.Header.Get("Content-Disposition")
	match := re.FindStringSubmatch(header)
	if len(match) > 1 {
		return match[1], nil
	}
	return "", nil
}

type progressBar struct {
	fileName     string
	fileSize     int64
	currentSize  int64
	width        int
	lastProgress int
}

func NewProgressBar(fileName string, fileSize int64) *progressBar {
	return &progressBar{
		fileName: fileName,
		fileSize: fileSize,
		width:    50,
	}
}

func (p *progressBar) Write(b []byte) (int, error) {
	n := len(b)
	p.currentSize += int64(n)
	progress := int(float64(p.currentSize*100) / float64(p.fileSize))

	if progress > p.lastProgress {
		p.lastProgress = progress

		content := fmt.Sprintf("Downloading: %s -------- %d%% (%s/%s)",
			p.fileName,
			progress,
			formatSize(p.currentSize),
			formatSize(p.fileSize),
		)

		PrintInline(content)
		if progress == 100 {
			fmt.Println()
		}
	}

	return n, nil
}
