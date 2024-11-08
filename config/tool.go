package config

import (
	"buildenv/pkg/io"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Tool struct {
	Url     string `json:"url"`
	RunPath string `json:"run_path"`
	Md5     string `json:"md5"`

	// Internal fields.
	toolName string `json:"-"`
}

func (t *Tool) Init(toolpath string) error {
	// Check if tool.json exists.
	if !pathExists(toolpath) {
		return fmt.Errorf("%s doesn't exists", toolpath)
	}

	// Read json file.
	bytes, err := os.ReadFile(toolpath)
	if err != nil {
		return fmt.Errorf("%s not exists", toolpath)
	}
	if err := json.Unmarshal(bytes, t); err != nil {
		return fmt.Errorf("%s is not valid: %w", toolpath, err)
	}

	// Set internal fields.
	t.toolName = strings.TrimSuffix(filepath.Base(toolpath), ".json")
	return nil
}

func (t *Tool) Verify(args VerifyArgs) error {
	if t.Url == "" {
		return fmt.Errorf("url of %s is empty", t.toolName)
	}

	if t.RunPath == "" {
		return fmt.Errorf("path of %s is empty", t.toolName)
	}

	if !args.CheckAndRepair {
		return nil
	}

	return t.checkAndRepair()
}

func (t Tool) checkAndRepair() error {
	toolPath := filepath.Join(Dirs.DownloadDir, t.RunPath)
	if pathExists(toolPath) {
		return nil
	}

	fileName := filepath.Base(t.Url)

	// Download to fixed dir.
	downloaded, err := io.Download(t.Url, Dirs.DownloadDir)
	if err != nil {
		return fmt.Errorf("%s: download failed: %w", fileName, err)
	}

	// Extract archive file.
	extractPath := filepath.Join(Dirs.DownloadDir, t.toolName)
	if err := io.Extract(downloaded, extractPath); err != nil {
		return fmt.Errorf("%s: extract failed: %w", fileName, err)
	}

	fmt.Printf("[✔] -------- %s.\n\n", fileName)
	return nil
}
