package io

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PathExists checks if the path exists.
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	return !os.IsNotExist(err)
}

// FileBaseName it's a improved version to get file base name.
func FileBaseName(fileName string) string {
	index := strings.Index(fileName, ".tar.")
	if index > 0 {
		return fileName[:index]
	}

	ext := filepath.Ext(fileName)
	return strings.TrimSuffix(fileName, ext)
}

func MoveNestedFolderIfExist(filePath string) error {
	// We assume the archive contains a single root folder, check if it has nested folder.
	if nestedFolder := findNestedFolder(filePath); nestedFolder != "" {
		// Move the entire nested folder to the parent directory
		if err := moveDirectoryToParent(nestedFolder, filepath.Dir(filePath)); err != nil {
			return err
		}
	}

	return nil
}

func findNestedFolder(parentDir string) string {
	entries, err := os.ReadDir(parentDir)
	if err != nil {
		return ""
	}

	folderName := filepath.Base(parentDir)

	for _, entry := range entries {
		// If a folder is found that isn't the one we are currently in,
		// it's considered a nested folder.
		if entry.IsDir() && folderName == entry.Name() {
			nestedDir := filepath.Join(parentDir, entry.Name())
			if _, err := os.Stat(nestedDir); err == nil {
				return nestedDir
			}
		}
	}

	return ""
}

func moveDirectoryToParent(nestedFolder, parentFolder string) error {
	destPath := filepath.Join(parentFolder, filepath.Base(nestedFolder))
	tmpPath := filepath.Join(parentFolder, filepath.Base(nestedFolder)+".tmp")

	// Move folder that we want to a temporary path.
	if err := os.Rename(nestedFolder, tmpPath); err != nil {
		return fmt.Errorf("failed to rename directory from %s to %s: %w", nestedFolder, nestedFolder+".old", err)
	}

	// Remove the now empty nested folder.
	if err := os.RemoveAll(destPath); err != nil {
		return fmt.Errorf("failed to remove empty nested folder %s: %w", nestedFolder, err)
	}

	// Convert the temporary folder to the actual folder.
	if err := os.Rename(tmpPath, destPath); err != nil {
		return fmt.Errorf("failed to move directory from %s to %s: %w", nestedFolder, destPath, err)
	}

	return nil
}
