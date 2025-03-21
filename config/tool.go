package config

import (
	"buildenv/pkg/color"
	"buildenv/pkg/fileio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Tool struct {
	Url         string `json:"url"`          // Download url.
	ArchiveName string `json:"archive_name"` // Archive name can be changed to avoid conflict.
	Path        string `json:"path"`         // Runtime path of tool, it's relative path  and would be converted to absolute path later.

	// Internal fields.
	toolName  string `json:"-"`
	fullpath  string `json:"-"`
	cmakepath string `json:"-"`
}

func (t *Tool) Init(toolpath string) error {
	// Check if tool.json exists.
	if !fileio.PathExists(toolpath) {
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

func (t *Tool) Validate() error {
	// Validate tool download url.
	if t.Url == "" {
		return fmt.Errorf("url of %s is empty", t.toolName)
	}

	// Validate tool path and convert to absolute path.
	if t.Path == "" {
		return fmt.Errorf("path of %s is empty", t.toolName)
	}

	t.fullpath = filepath.Join(Dirs.ExtractedToolsDir, t.Path)
	t.cmakepath = fmt.Sprintf("${BUILDENV_ROOT_DIR}/downloads/tools/%s", t.Path)

	// This is used to cross-compile other ports by buildenv.
	os.Setenv("PATH", t.fullpath+string(os.PathListSeparator)+os.Getenv("PATH"))

	return nil
}

func (t Tool) CheckAndRepair(args SetupArgs) error {
	if !args.RepairBuildenv() {
		return nil
	}

	// Default folder name would be the first folder of path,
	// it also can be specified by archiveName.
	folderName := strings.Split(t.Path, string(filepath.Separator))[0]
	if t.ArchiveName != "" {
		folderName = fileio.FileBaseName(t.ArchiveName)
	}

	location := filepath.Join(Dirs.ExtractedToolsDir, folderName)

	// Check if tool exists.
	if fileio.PathExists(t.fullpath) {
		if !args.Silent() {
			title := color.Sprintf(color.Green, "\n[✔] ---- Tool: %s\n", fileio.FileBaseName(t.Url))
			fmt.Printf("%sLocation: %s\n", title, location)
		}
		return nil
	}

	// Use archive name as download file name if specified.
	archiveName := filepath.Base(t.Url)
	if t.ArchiveName != "" {
		archiveName = t.ArchiveName
	}

	// Check and repair resource.
	repair := fileio.NewDownloadRepair(t.Url, archiveName, folderName, Dirs.ExtractedToolsDir, Dirs.DownloadedDir)
	if err := repair.CheckAndRepair(); err != nil {
		return err
	}

	// Print download & extract info.
	if !args.Silent() {
		title := color.Sprintf(color.Green, "\n[✔] ---- Tool: %s\n", fileio.FileBaseName(t.Url))
		fmt.Printf("%sLocation: %s\n", title, location)
	}
	return nil
}

func (t Tool) Write(toolPath string) error {
	bytes, err := json.MarshalIndent(t, "", "    ")
	if err != nil {
		return err
	}

	// Check if tool exists.
	if fileio.PathExists(toolPath) {
		return fmt.Errorf("%s is already exists", toolPath)
	}

	// Make sure the parent directory exists.
	parentDir := filepath.Dir(toolPath)
	if err := os.MkdirAll(parentDir, os.ModeDir|os.ModePerm); err != nil {
		return err
	}
	return os.WriteFile(toolPath, bytes, os.ModePerm)
}
