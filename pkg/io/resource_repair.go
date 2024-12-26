package io

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func NewResourceRepair(url, archiveName, folderName, extractedToolsDir, downloadRootDir string) *ResourceRepair {
	return &ResourceRepair{
		url:               url,
		archiveName:       archiveName,
		folderName:        folderName,
		extractedToolsDir: extractedToolsDir,
		downloadRootDir:   downloadRootDir,
	}
}

type ResourceRepair struct {
	url               string
	archiveName       string
	folderName        string
	extractedToolsDir string
	downloadRootDir   string
}

func (r ResourceRepair) CheckAndRepair() error {
	switch {
	case strings.HasPrefix(r.url, "http"), strings.HasPrefix(r.url, "ftp"):
		downloaded, err := r.download(r.url, r.archiveName)
		if err != nil {
			return nil
		}

		// Extract archive file.
		if err := Extract(downloaded, filepath.Join(r.extractedToolsDir, r.folderName)); err != nil {
			return fmt.Errorf("%s: extract failed: %w", downloaded, err)
		}

		// Check if has nested folder (handling case where there's an extra nested folder).
		extractedPath := filepath.Join(r.extractedToolsDir, r.folderName)
		if err := MoveNestedFolderIfExist(extractedPath); err != nil {
			return fmt.Errorf("%s: failed to move nested folder: %w", r.folderName, err)
		}

	case strings.HasPrefix(r.url, "file:///"):
		localPath := strings.TrimPrefix(r.url, "file:///")
		state, err := os.Stat(localPath)
		if err != nil {
			return fmt.Errorf("%s is not accessable", r.url)
		}

		// If localPath is a directory, we assume it is valid.
		if state.IsDir() {
			return nil
		}

		// Extract archive file.
		if err := Extract(localPath, filepath.Join(r.extractedToolsDir, r.folderName)); err != nil {
			return fmt.Errorf("%s: extract failed: %w", localPath, err)
		}

		// Check if has nested folder (handling case where there's an extra nested folder).
		extractedPath := filepath.Join(r.extractedToolsDir, r.folderName)
		if err := MoveNestedFolderIfExist(extractedPath); err != nil {
			return fmt.Errorf("%s: failed to move nested folder: %w", r.folderName, err)
		}

	default:
		return fmt.Errorf("%s is not accessible", r.url)
	}

	return nil
}

func (r ResourceRepair) download(url, archiveName string) (downloaded string, err error) {
	downloaded = filepath.Join(r.downloadRootDir, archiveName)
	if PathExists(downloaded) {
		// Redownload if remote file size and local file size not match.
		fileSize, err := FileSize(url)
		if err != nil {
			return "", fmt.Errorf("%s: get remote filesize failed: %w", archiveName, err)
		}
		info, err := os.Stat(downloaded)
		if err != nil {
			return "", fmt.Errorf("%s: get local filesize failed: %w", archiveName, err)
		}
		if info.Size() != fileSize {
			downloadRequest := NewDownloadRequest(url, r.downloadRootDir)
			downloadRequest.SetArchiveName(archiveName)
			if _, err := downloadRequest.Download(); err != nil {
				return "", fmt.Errorf("%s: download failed: %w", archiveName, err)
			}
		}
	} else {
		downloadRequest := NewDownloadRequest(url, r.downloadRootDir)
		downloadRequest.SetArchiveName(archiveName)
		if _, err := downloadRequest.Download(); err != nil {
			return "", fmt.Errorf("%s: download failed: %w", archiveName, err)
		}
	}

	return downloaded, nil
}
