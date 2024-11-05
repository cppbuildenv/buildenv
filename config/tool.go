package config

import (
	"buildenv/pkg/io"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Tool struct {
	Url         string `json:"url"`
	RunPath     string `json:"run_path"`
	ExtractPath string `json:"extract_path"`
	Md5         string `json:"md5"`
	ToolName    string `json:"-"`
}

func (t *Tool) Read(toolpath string) error {
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

	t.ToolName = filepath.Base(toolpath)
	return nil
}

func (t *Tool) Verify(checkAndRepiar bool) error {
	if t.Url == "" {
		return fmt.Errorf("url of %s is empty", t.ToolName)
	}

	if t.RunPath == "" {
		return fmt.Errorf("path of %s is empty", t.ToolName)
	}

	if !checkAndRepiar {
		return nil
	}

	return t.checkAndRepair()
}

func (t Tool) checkAndRepair() error {
	toolPath := filepath.Join(WorkspaceDir, t.RunPath)
	if pathExists(toolPath) {
		return nil
	}

	fileName := filepath.Base(t.Url)

	// Download to fixed dir.
	downloaded, err := io.Download(t.Url, DownloadDir)
	if err != nil {
		return fmt.Errorf("%s: download failed: %w", fileName, err)
	}

	// Extract to `extract_path`.
	extractDir := filepath.Join(WorkspaceDir, t.ExtractPath)
	if err := io.Extract(downloaded, extractDir); err != nil {
		return fmt.Errorf("%s: extract failed: %w", fileName, err)
	}

	fmt.Printf("[✔] -------- %s.\n\n", fileName)
	return nil
}
