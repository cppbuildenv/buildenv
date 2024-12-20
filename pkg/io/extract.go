package io

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func Extract(archiveFile, destDir string) error {
	fileName := filepath.Base(archiveFile)
	PrintInline(fmt.Sprintf("\rExtracting:  %s...", fileName))

	var command string

	switch {
	case strings.HasSuffix(archiveFile, ".tar.gz"):
		command = fmt.Sprintf("tar -zxvf %s -C %s", archiveFile, destDir)

	case strings.HasSuffix(archiveFile, ".tar.xz"):
		command = fmt.Sprintf("tar -xf %s -C %s", archiveFile, destDir)

	case strings.HasSuffix(archiveFile, ".tar.bz2"):
		command = fmt.Sprintf("tar -xvjf %s -C %s", archiveFile, destDir)

	case strings.HasSuffix(archiveFile, ".zip"):
		command = fmt.Sprintf("unzip %s -d %s", archiveFile, destDir)

	case strings.HasSuffix(archiveFile, ".7z"):
		command = fmt.Sprintf("7z x %s -o %s", archiveFile, destDir)

	default:
		return fmt.Errorf("unsupported archive file type: %s", archiveFile)
	}

	if err := os.MkdirAll(destDir, os.ModeDir|os.ModePerm); err != nil {
		return fmt.Errorf("failed to mkdir for extract: %w", err)
	}

	// Run command.
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("bash", "-c", command)
	}

	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract: %w", err)
	}

	fmt.Println()
	return nil
}
