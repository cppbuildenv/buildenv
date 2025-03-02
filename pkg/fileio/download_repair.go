package fileio

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func NewDownloadRepair(url, archiveName, folderName, extractTo, downloadedDir string) *DownloadRepair {
	return &DownloadRepair{
		url:           url,
		archiveName:   archiveName,
		folderName:    folderName,
		extractTo:     extractTo,
		downloadedDir: downloadedDir,
	}
}

type DownloadRepair struct {
	url           string
	archiveName   string
	folderName    string
	extractTo     string
	downloadedDir string
}

func (d DownloadRepair) CheckAndRepair() error {
	switch {
	case strings.HasPrefix(d.url, "http"), strings.HasPrefix(d.url, "ftp"):
		downloaded, err := d.download(d.url, d.archiveName)
		if err != nil {
			return err
		}

		// Extract archive file.
		if err := Extract(downloaded, filepath.Join(d.extractTo, d.folderName)); err != nil {
			return fmt.Errorf("%s: extract failed: %w", downloaded, err)
		}

		// Check if has nested folder (handling case where there's an nested folder).
		extractedPath := filepath.Join(d.extractTo, d.folderName)
		if err := moveNestedFolderIfExist(extractedPath); err != nil {
			return fmt.Errorf("%s: failed to move nested folder: %w", d.folderName, err)
		}

	case strings.HasPrefix(d.url, "file:///"):
		localPath := strings.TrimPrefix(d.url, "file:///")
		state, err := os.Stat(localPath)
		if err != nil {
			return fmt.Errorf("%s is not accessable", d.url)
		}

		// If localPath is a directory, we assume it is valid.
		if state.IsDir() {
			return nil
		}

		// Extract archive file.
		if err := Extract(localPath, filepath.Join(d.extractTo, d.folderName)); err != nil {
			return fmt.Errorf("%s: extract failed: %w", localPath, err)
		}

		// Check if has nested folder (handling case where there's an extra nested folder).
		extractedPath := filepath.Join(d.extractTo, d.folderName)
		if err := moveNestedFolderIfExist(extractedPath); err != nil {
			return fmt.Errorf("%s: failed to move nested folder: %w", d.folderName, err)
		}

	default:
		return fmt.Errorf("%s is not accessible", d.url)
	}

	return nil
}

func (d *DownloadRepair) MoveAllToParent() error {
	entries, err := os.ReadDir(filepath.Join(d.extractTo, d.folderName))
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if err := RenameFile(entry.Name(), d.extractTo); err != nil {
			return err
		}
	}

	return nil
}

func (d DownloadRepair) download(url, archiveName string) (downloaded string, err error) {
	downloaded = filepath.Join(d.downloadedDir, archiveName)
	if PathExists(downloaded) {
		// Redownload if remote file size and local file size not match.
		fileSize, err := FileSize(url)
		if err != nil || fileSize <= 0 {
			return "", fmt.Errorf("%s: get remote filesize failed: %w", archiveName, err)
		}
		info, err := os.Stat(downloaded)
		if err != nil {
			return "", fmt.Errorf("%s: get local filesize failed: %w", archiveName, err)
		}
		if info.Size() != fileSize {
			downloadRequest := NewDownloadRequest(url, d.downloadedDir)
			downloadRequest.SetArchiveName(archiveName)
			if _, err := downloadRequest.Download(); err != nil {
				return "", fmt.Errorf("%s: download failed: %w", archiveName, err)
			}
		}
	} else {
		downloadRequest := NewDownloadRequest(url, d.downloadedDir)
		downloadRequest.SetArchiveName(archiveName)
		if _, err := downloadRequest.Download(); err != nil {
			return "", fmt.Errorf("%s: download failed: %w", archiveName, err)
		}
	}

	return downloaded, nil
}
