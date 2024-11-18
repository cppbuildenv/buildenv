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

	// Internal fields.
	toolName string `json:"-"`
	fullpath string `json:"-"`
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

	// Verify tool download url.
	if t.Url == "" {
		return fmt.Errorf("url of %s is empty", t.toolName)
	}
	if err := io.CheckAvailable(t.Url); err != nil {
		return fmt.Errorf("tool.url of %s is not accessible: %w", t.toolName, err)
	}

	// Verify tool path and convert to absolute path.
	if t.Path == "" {
		return fmt.Errorf("path of %s is empty", t.toolName)
	}
	toolPath, err := toAbsPath(t.Path)
	if err != nil {
		return fmt.Errorf("cannot get absolute path: %s", t.Path)
	}
	t.fullpath = toolPath

	// This is used to cross-compile other ports by buildenv.
	os.Setenv("PATH", fmt.Sprintf("%s:%s", t.fullpath, os.Getenv("PATH")))

	if !args.CheckAndRepair {
		return nil
	}

	return t.checkAndRepair()
}

func (t Tool) checkAndRepair() error {
	// Check if tool exists.
	if io.PathExists(t.fullpath) {
		return nil
	}

	// Check if need to download file.
	fileName := filepath.Base(t.Url)
	downloaded := filepath.Join(Dirs.DownloadRootDir, fileName)
	if io.PathExists(downloaded) {
		// Redownload if remote file size and local file size not match.
		fileSize, err := io.FileSize(t.Url)
		if err != nil {
			return fmt.Errorf("%s: get remote filesize failed: %w", fileName, err)
		}
		info, err := os.Stat(downloaded)
		if err != nil {
			return fmt.Errorf("%s: get local filesize failed: %w", fileName, err)
		}
		if info.Size() != fileSize {
			if _, err := io.Download(t.Url, Dirs.DownloadRootDir); err != nil {
				return fmt.Errorf("%s: download failed: %w", fileName, err)
			}
		}
	} else {
		if _, err := io.Download(t.Url, Dirs.DownloadRootDir); err != nil {
			return fmt.Errorf("%s: download failed: %w", fileName, err)
		}
	}

	// Extract archive file.
	folderName := strings.Split(t.Path, string(filepath.Separator))[0]
	if err := io.Extract(downloaded, filepath.Join(Dirs.DownloadRootDir, folderName)); err != nil {
		return fmt.Errorf("%s: extract failed: %w", fileName, err)
	}

	// Check if has nested folder (handling case where there's an extra nested folder).
	extractPath := filepath.Join(Dirs.DownloadRootDir, folderName)
	if err := io.MoveNestedFolderIfExist(extractPath); err != nil {
		return fmt.Errorf("%s: failed to move nested folder: %w", fileName, err)
	}

	// Print download & extract info.
	fmt.Print(color.Sprintf(color.Blue, "[✔] -------- %s (tool: %s)\n\n", fileName, extractPath))
	return nil
}
