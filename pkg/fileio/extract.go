package fileio

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func IsSupportedArchive(filePath string) bool {
	return strings.HasSuffix(filePath, ".tar.gz") ||
		strings.HasSuffix(filePath, ".tar.xz") ||
		strings.HasSuffix(filePath, ".tar.bz2") ||
		strings.HasSuffix(filePath, ".zip") ||
		strings.HasSuffix(filePath, ".7z")
}

func Extract(archiveFile, destDir string) error {
	fileName := filepath.Base(archiveFile)
	PrintInline(fmt.Sprintf("\rExtracting: %s...", fileName))

	var command string

	switch {
	case strings.HasSuffix(archiveFile, ".tar.gz"), strings.HasSuffix(archiveFile, ".tgz"):
		command = fmt.Sprintf("tar -zxvf %s -C %s", archiveFile, destDir)

	case strings.HasSuffix(archiveFile, ".tar.xz"):
		command = fmt.Sprintf("tar -xvf %s -C %s", archiveFile, destDir)

	case strings.HasSuffix(archiveFile, ".tar.bz2"):
		command = fmt.Sprintf("tar -xvjf %s -C %s", archiveFile, destDir)

	case strings.HasSuffix(archiveFile, ".zip"):
		command = fmt.Sprintf("unzip %s -d %s", archiveFile, destDir)

	case strings.HasSuffix(archiveFile, ".7z"):
		command = fmt.Sprintf("7z x %s -o %s", archiveFile, destDir)

	default:
		return fmt.Errorf("unsupported archive file type: %s", archiveFile)
	}

	if err := os.RemoveAll(destDir); err != nil {
		return fmt.Errorf("failed to remove directory: %w", err)
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

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract: %w", err)
	}

	return nil
}

// Targz creates a tarball from srcDir and saves it to archivePath.
func Targz(archivePath, srcDir string, includeFolder bool) error {
	var cmd *exec.Cmd
	var command string

	if includeFolder {
		command = fmt.Sprintf("tar -cvzf %s %s", archivePath, srcDir)
	} else {
		command = fmt.Sprintf("tar -cvzf %s -C %s .", archivePath, srcDir)
	}

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("bash", "-c", command)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create tarball: %w", err)
	}

	return nil
}
