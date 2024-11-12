package deploy

import (
	"buildenv/pkg/io"
	"fmt"
	"path/filepath"
	"strings"
)

type DeployConfig struct {
	InstalledDir string `json:"-"`
	DownloadDir  string `json:"-"`
}

func (d DeployConfig) CheckAndRepair(url string) error {
	// Download to fixed dir.
	downloaded, err := io.Download(url, d.DownloadDir)
	if err != nil {
		return fmt.Errorf("%s: download port failed: %w", url, err)
	}

	// Extract archive file.
	fileName := filepath.Base(url)
	folderName := strings.TrimSuffix(fileName, ".tar.gz")
	extractPath := filepath.Join(d.InstalledDir, folderName)
	if err := io.Extract(downloaded, extractPath); err != nil {
		return fmt.Errorf("%s: extract %s failed: %w", fileName, downloaded, err)
	}

	return nil
}
