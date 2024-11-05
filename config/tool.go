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
	RuntimePath string `json:"runtime_path"`
	ExtractPath string `json:"extract_path"`
	Md5         string `json:"md5"`
	ToolName    string `json:"-"`
}

func (t *Tool) Read(toolName string) error {
	// Check if tool.json exists.
	toolPath := filepath.Join(ToolsDir, toolName+".json")
	if !pathExists(toolPath) {
		return fmt.Errorf("config file of %s doesn't exists in %s", toolName, ToolsDir)
	}

	// Read json file.
	bytes, err := os.ReadFile(toolPath)
	if err != nil {
		return fmt.Errorf("config file of %s not exists in %s", toolName, ToolsDir)
	}
	if err := json.Unmarshal(bytes, t); err != nil {
		return fmt.Errorf("config file of %s is not valid: %w", toolName, err)
	}

	t.ToolName = toolName
	return nil
}

func (t *Tool) Verify(checkAndRepiar bool) error {
	if t.Url == "" {
		return fmt.Errorf("url of %s is empty", t.ToolName)
	}

	if t.RuntimePath == "" {
		return fmt.Errorf("path of %s is empty", t.ToolName)
	}

	if !checkAndRepiar {
		return nil
	}

	return t.checkAndRepair()
}

func (t Tool) checkAndRepair() error {
	toolPath := filepath.Join(WorkspaceDir, t.RuntimePath)
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
