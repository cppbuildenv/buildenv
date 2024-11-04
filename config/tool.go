package config

import (
	"buildenv/pkg/io"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
)

type Tool struct {
	Url  string `json:"url"`
	Path string `json:"path"`
	Md5  string `json:"md5"`
}

func (t *Tool) Verify(resRepoUrl string, toolName string, onlyFields bool) error {
	// Check if tool.json exists.
	toolPath := filepath.Join(ToolsDir, toolName+".json")
	if !pathExists(toolPath) {
		return fmt.Errorf("config file of %s is not exists in %s", toolName, ToolsDir)
	}

	// Read json file.
	bytes, err := os.ReadFile(toolPath)
	if err != nil {
		return fmt.Errorf("config file of %s not exists in %s", toolName, ToolsDir)
	}
	if err := json.Unmarshal(bytes, t); err != nil {
		return fmt.Errorf("config file of %s is not valid: %w", toolName, err)
	}

	if t.Url == "" {
		return fmt.Errorf("url of %s is empty", toolName)
	}

	if t.Path == "" {
		return fmt.Errorf("path of %s is empty", toolName)
	}

	if onlyFields {
		return nil
	}

	return t.ensureIntegrity(resRepoUrl)
}

func (t Tool) ensureIntegrity(resRepoUrl string) error {
	toolPath := filepath.Join(WorkspaceDir, t.Path)
	if pathExists(toolPath) {
		return nil
	}

	fullUrl, err := url.JoinPath(resRepoUrl, t.Url)
	if err != nil {
		return err
	}

	fileName := filepath.Base(t.Url)

	// Download to fixed dir.
	downloaded, err := io.Download(fullUrl, DownloadDir)
	if err != nil {
		return fmt.Errorf("%s: download failed: %w", fileName, err)
	}

	// Extract to dir with same parent.
	parentDir := filepath.Dir(t.Url)
	extractDir := filepath.Join(WorkspaceDir, parentDir)
	if err := io.Extract(downloaded, extractDir); err != nil {
		return fmt.Errorf("%s: extract failed: %w", fileName, err)
	}

	fmt.Printf("[✔] ---- %s of platform setup success.\n\n", fileName)
	return nil
}
