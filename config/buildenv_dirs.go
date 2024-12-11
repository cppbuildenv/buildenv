package config

import (
	"fmt"
	"os"
	"path/filepath"
)

var Dirs = newDirs()

type dirs struct {
	WorkspaceDir      string // absolute path of "."
	DownloadRootDir   string // absolute path of "downloads"
	ExtractedToolsDir string // absolute path of "downloaded/tools"
	InstalledRootDir  string // absolute path of "installed"
	PlatformDir       string // absolute path of "conf/platforms"
	ProjectDir        string // absolute path of "conf/projects"
	ToolDir           string // absolute path of "conf/tools"
	PortDir           string // absolute path of "conf/ports"
}

func newDirs() *dirs {
	var dirs dirs
	// Set current dir as workspaceDir.
	currentDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("get current dir failed: %s", err.Error()))
	}

	// Set default paths.
	dirs.WorkspaceDir = currentDir
	dirs.PlatformDir = filepath.Join(dirs.WorkspaceDir, "conf", "platforms")
	dirs.ProjectDir = filepath.Join(dirs.WorkspaceDir, "conf", "projects")
	dirs.DownloadRootDir = filepath.Join(dirs.WorkspaceDir, "downloads")
	dirs.ExtractedToolsDir = filepath.Join(dirs.WorkspaceDir, "downloads", "tools")
	dirs.InstalledRootDir = filepath.Join(dirs.WorkspaceDir, "installed")
	dirs.ToolDir = filepath.Join(dirs.WorkspaceDir, "conf", "tools")
	dirs.PortDir = filepath.Join(dirs.WorkspaceDir, "conf", "ports")

	return &dirs
}
