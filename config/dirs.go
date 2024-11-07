package config

import (
	"fmt"
	"os"
	"path/filepath"
)

var Dirs = newDirs()

type dirs struct {
	WorkspaceDir string // Default rootfs is "."
	DownloadDir  string // Default downloadDir is "downloads"
	InstalledDir string // Default installedDir is "installed"
	PlatformDir  string // Default platformDir is "conf/platforms"
	ToolDir      string // Default toolDir is "conf/tools"
	PortDir      string // Default portDir is "conf/ports"
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
	dirs.DownloadDir = filepath.Join(dirs.WorkspaceDir, "downloads")
	dirs.InstalledDir = filepath.Join(dirs.WorkspaceDir, "installed")
	dirs.ToolDir = filepath.Join(dirs.WorkspaceDir, "conf", "tools")
	dirs.PortDir = filepath.Join(dirs.WorkspaceDir, "conf", "ports")

	return &dirs
}
