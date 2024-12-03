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
	Url         string `json:"url"`                    // Download url.
	ArchiveName string `json:"archive_name,omitempty"` // Archive name can be changed to avoid conflict.
	Path        string `json:"path"`                   // Runtime path of tool, it's relative path  and would be converted to absolute path later.

	// Internal fields.
	toolName  string `json:"-"`
	fullpath  string `json:"-"`
	cmakepath string `json:"-"`
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

func (t *Tool) Verify() error {
	// Verify tool download url.
	if t.Url == "" {
		return fmt.Errorf("url of %s is empty", t.toolName)
	}
	if err := io.CheckAvailable(t.Url); err != nil {
		return fmt.Errorf("tool.url of %s is not accessible", t.Url)
	}

	// Verify tool path and convert to absolute path.
	if t.Path == "" {
		return fmt.Errorf("path of %s is empty", t.toolName)
	}

	t.fullpath = filepath.Join(Dirs.ExtractedToolsDir, t.Path)
	t.cmakepath = fmt.Sprintf("${BUILDENV_ROOT_DIR}/downloads/tools/%s", t.Path)

	// This is used to cross-compile other ports by buildenv.
	os.Setenv("PATH", fmt.Sprintf("%s:%s", t.fullpath, os.Getenv("PATH")))

	return nil
}

func (t Tool) CheckAndRepair(args VerifyArgs) error {
	if !args.CheckAndRepair() {
		return nil
	}

	// Default folder name would be the first folder of path,
	// it also can be specified by archiveName.
	folderName := strings.Split(t.Path, string(filepath.Separator))[0]
	if t.ArchiveName != "" {
		folderName = io.FileBaseName(t.ArchiveName)
	}

	extractPath := filepath.Join(Dirs.ExtractedToolsDir, folderName)

	// Check if tool exists.
	if io.PathExists(t.fullpath) {
		// No need to show rootfs state info when install a port.
		if args.PortToInstall() == "" && !args.Silent() {
			fmt.Print(color.Sprintf(color.Blue, "[✔] -------- Tool: %s\nLocation: %s\n\n",
				io.FileBaseName(t.Url), extractPath))
		}
		return nil
	}

	// Use archive name as download file name if specified.
	archiveName := filepath.Base(t.Url)
	if t.ArchiveName != "" {
		archiveName = t.ArchiveName
	}

	// Check if need to download file.
	downloaded := filepath.Join(Dirs.DownloadRootDir, archiveName)
	if io.PathExists(downloaded) {
		// Redownload if remote file size and local file size not match.
		fileSize, err := io.FileSize(t.Url)
		if err != nil {
			return fmt.Errorf("%s: get remote filesize failed: %w", archiveName, err)
		}
		info, err := os.Stat(downloaded)
		if err != nil {
			return fmt.Errorf("%s: get local filesize failed: %w", archiveName, err)
		}
		if info.Size() != fileSize {
			downloadRequest := io.NewDownloadRequest(t.Url, Dirs.DownloadRootDir)
			downloadRequest.SetArchiveName(archiveName)
			if _, err := downloadRequest.Download(); err != nil {
				return fmt.Errorf("%s: download failed: %w", archiveName, err)
			}
		}
	} else {
		downloadRequest := io.NewDownloadRequest(t.Url, Dirs.DownloadRootDir)
		downloadRequest.SetArchiveName(archiveName)
		if _, err := downloadRequest.Download(); err != nil {
			return fmt.Errorf("%s: download failed: %w", archiveName, err)
		}
	}

	// Extract archive file.
	if err := io.Extract(downloaded, filepath.Join(Dirs.ExtractedToolsDir, folderName)); err != nil {
		return fmt.Errorf("%s: extract failed: %w", archiveName, err)
	}

	// Check if has nested folder (handling case where there's an extra nested folder).
	if err := io.MoveNestedFolderIfExist(extractPath); err != nil {
		return fmt.Errorf("%s: failed to move nested folder: %w", archiveName, err)
	}

	// Print download & extract info.
	if !args.Silent() {
		fmt.Print(color.Sprintf(color.Blue, "[✔] -------- Tool: %s\nLocation: %s\n\n",
			io.FileBaseName(t.Url), extractPath))
	}
	return nil
}
