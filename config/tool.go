package config

import (
	"buildenv/pkg/color"
	"buildenv/pkg/io"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Tool struct {
	Url  string `json:"url"`
	Path string `json:"path"`
	Md5  string `json:"md5"`

	// Internal fields.
	toolName string `json:"-"`
}

func (t *Tool) Init(toolpath string) error {
	// Check if tool.json exists.
	if !io.PathExists(toolpath) {
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
	// Relative path -> Absolute path.
	var toAbsPath = func(relativePath string) (string, error) {
		path := filepath.Join(Dirs.DownloadRootDir, relativePath)
		rootfsPath, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}

		return rootfsPath, nil
	}

	if t.Url == "" {
		return fmt.Errorf("url of %s is empty", t.toolName)
	}

	// Verify tool path and convert to absolute path.
	if t.Path == "" {
		return fmt.Errorf("path of %s is empty", t.toolName)
	}
	toolPath, err := toAbsPath(t.Path)
	if err != nil {
		return fmt.Errorf("cannot get absolute path: %s", t.Path)
	}
	t.Path = toolPath

	// This is used to cross-compile other ports by buildenv.
	os.Setenv("PATH", fmt.Sprintf("%s:%s", t.Path, os.Getenv("PATH")))

	if !args.CheckAndRepair {
		return nil
	}

	return t.checkAndRepair()
}

func (t Tool) checkAndRepair() error {
	// Check if tool exists.
	if io.PathExists(t.Path) {
		return nil
	}

	fileName := filepath.Base(t.Url)

	// Download to fixed dir.
	downloaded, err := io.Download(t.Url, Dirs.DownloadRootDir)
	if err != nil {
		return fmt.Errorf("%s: download failed: %w", fileName, err)
	}

	// Extract archive file.
	extractPath := filepath.Join(Dirs.DownloadRootDir, t.toolName)
	if err := io.Extract(downloaded, extractPath); err != nil {
		return fmt.Errorf("%s: extract failed: %w", fileName, err)
	}

	fmt.Print(color.Sprintf(color.Blue, "[✔] -------- %s (tool: %s)\n\n", fileName, extractPath))
	return nil
}
