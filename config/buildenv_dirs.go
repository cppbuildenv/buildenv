package config

import (
	"fmt"
	"os"
	"path/filepath"
)

var Dirs = newDirs()

type dirs struct {
	WorkspaceDir      string // absolute path of "."
	PlatformsDir      string // absolute path of "conf/platforms"
	ProjectsDir       string // absolute path of "conf/projects"
	PackagesDir       string // absolute path of "packages"
	DownloadedDir     string // absolute path of "downloads"
	ExtractedToolsDir string // absolute path of "downloaded/tools"
	InstalledDir      string // absolute path of "installed"
	ToolsDir          string // absolute path of "conf/tools"
	PortsDir          string // absolute path of "conf/ports"
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
	dirs.PlatformsDir = filepath.Join(dirs.WorkspaceDir, "conf", "platforms")
	dirs.ProjectsDir = filepath.Join(dirs.WorkspaceDir, "conf", "projects")
	dirs.PackagesDir = filepath.Join(dirs.WorkspaceDir, "packages")
	dirs.DownloadedDir = filepath.Join(dirs.WorkspaceDir, "downloads")
	dirs.ExtractedToolsDir = filepath.Join(dirs.WorkspaceDir, "downloads", "tools")
	dirs.InstalledDir = filepath.Join(dirs.WorkspaceDir, "installed")
	dirs.ToolsDir = filepath.Join(dirs.WorkspaceDir, "conf", "tools")
	dirs.PortsDir = filepath.Join(dirs.WorkspaceDir, "conf", "ports")

	return &dirs
}
