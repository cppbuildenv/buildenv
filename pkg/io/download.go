package io

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func Download(url string, destDir string, progress func(percent int)) (downloaded string, err error) {
	// Get file from url.
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to access file download url: %v", err)
	}
	defer resp.Body.Close()

	// Get file info from http header.
	fileInfo, err := getFileInfo(url)
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %v", err)
	}

	// Mkdir parent folders if not exist.
	if err := os.MkdirAll(destDir, os.ModeDir|os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create dir: %v", err)
	}

	// Create local file
	downloaded = filepath.Join(destDir, fileInfo.Name)
	outputFile, err := os.Create(downloaded)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer outputFile.Close()

	// Write data to file with progress.
	fileWriter := NewFileWriter(fileInfo.Name, fileInfo.Size, progress)
	if err := fileWriter.Write(resp.Body, destDir); err != nil {
		return "", fmt.Errorf("failed to write file: %v", err)
	}

	return downloaded, nil
}

type fileInfo struct {
	Name string
	Size int64
	Ext  string
}

func getFileInfo(url string) (fileInfo, error) {
	resp, err := http.Head(url)
	if err != nil {
		return fileInfo{}, err
	}
	defer resp.Body.Close()

	var filename string
	if value := resp.Header.Get("Content-Disposition"); value != "" {
		if parts := strings.Split(value, "filename="); len(parts) > 1 {
			filename = strings.Trim(parts[1], "\"")
		}
	} else {
		parts := strings.Split(url, "/")
		filename = parts[len(parts)-1]
	}

	return fileInfo{
		Name: filename,
		Size: resp.ContentLength,
	}, nil
}
