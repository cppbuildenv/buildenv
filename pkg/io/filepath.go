package io

import (
	"os"
	"path/filepath"
)

// PathExists checks if the path exists.
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	return !os.IsNotExist(err)
}

// ToAbsPath converts relative path to absolute path.
func ToAbsPath(parentPath, relativePath string) (string, error) {
	path := filepath.Join(parentPath, relativePath)
	rootfsPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	return rootfsPath, nil
}
