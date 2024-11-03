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
	"strings"
)

func Download(url string, destDir string) (downloaded string, err error) {
	// Read file size.
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	fileName, err := getFileName(url)
	if err != nil {
		return "", err
	}

	fileSize := resp.ContentLength
	progress := NewProgressBar(fileName, fileSize)

	if err := os.MkdirAll(destDir, os.ModeDir|os.ModePerm); err != nil {
		return "", err
	}

	// Build output filePath
	outputFile := filepath.Join(destDir, fileName)
	file, err := os.Create(outputFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Copy to local file with progress.
	_, err = io.Copy(io.MultiWriter(file, progress), resp.Body)
	if err != nil {
		return "", err
	}

	return outputFile, nil
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

		output := fmt.Sprintf("\rDownloading:\t%s ---- %d%% (%s/%s)",
			p.fileName,
			progress,
			formatSize(p.currentSize),
			formatSize(p.fileSize))

		// Add padding to align the output with the terminal width.
		padding := terminalWidth() - len(output) - 10
		if padding > 0 {
			output += strings.Repeat(" ", padding)
		}
		fmt.Printf("\r%s", output)

		if progress == 100 {
			fmt.Println()
		}
	}

	return n, nil
}
