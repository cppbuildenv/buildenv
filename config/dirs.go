package config

import (
	"fmt"
	"os"
	"path/filepath"
)

var Dirs = newDirs()

type dirs struct {
	WorkspaceDir     string // Default is "."
	DownloadRootDir  string // Default is "downloads"
	InstalledRootDir string // Default is "installed"
	PlatformDir      string // Default is "conf/platforms"
	ToolDir          string // Default is "conf/tools"
	PortDir          string // Default is "conf/ports"
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
	dirs.DownloadRootDir = filepath.Join(dirs.WorkspaceDir, "downloads")
	dirs.InstalledRootDir = filepath.Join(dirs.WorkspaceDir, "installed")
	dirs.ToolDir = filepath.Join(dirs.WorkspaceDir, "conf", "tools")
	dirs.PortDir = filepath.Join(dirs.WorkspaceDir, "conf", "ports")

	return &dirs
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	return !os.IsNotExist(err)
}
