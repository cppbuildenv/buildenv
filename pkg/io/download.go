package io

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
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
