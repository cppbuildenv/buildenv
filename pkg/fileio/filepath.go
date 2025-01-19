package fileio

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// IsExecutable check if file was executable
func IsExecutable(filepath string) bool {
	info, err := os.Stat(filepath)
	if err != nil {
		panic("file not found for " + filepath)
	}

	// 73: 000 001 001 001
	perm := info.Mode().Perm()
	flag := perm & os.FileMode(73)
	return uint32(flag) == uint32(73)
}

// IsReadable check if file or dir readable
func IsReadable(filepath string) bool {
	info, err := os.Stat(filepath)
	if err != nil {
		return false
	}

	return info.Mode().Perm()&(1<<(uint(8))) != 0
}

// IsWritable check if file or dir writable
func IsWritable(filepath string) bool {
	info, err := os.Stat(filepath)
	if err != nil {
		return false
	}

	return info.Mode().Perm()&(1<<(uint(7))) != 0
}

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
	fileName = filepath.Base(fileName)
	index := strings.Index(fileName, ".tar.")
	if index > 0 {
		return fileName[:index]
	}

	ext := filepath.Ext(fileName)
	return strings.TrimSuffix(fileName, ext)
}

// CopyFile copy file from src to dest.
func CopyFile(src, dest string) error {
	// Read file info.
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}

	// Create symlink if it's a symlink.
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(src)
		if err != nil {
			return err
		}

		// Remove dest if it exists before creating symlink.
		if _, err := os.Lstat(dest); err == nil {
			if err := os.Remove(dest); err != nil {
				return err
			}
		}

		return os.Symlink(target, dest)
	}

	// Copy normal file.
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return err
	}

	return nil
}

func RenameFile(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source: %v", err)
	}

	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(src)
		if err != nil {
			return fmt.Errorf("failed to read symlink: %v", err)
		}

		if err := os.MkdirAll(filepath.Dir(dst), os.ModeDir|os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory for renaming symlink: %v", err)
		}

		if err := os.Symlink(target, dst); err != nil {
			return fmt.Errorf("failed to create symlink: %v", err)
		}
	} else {
		if err := os.MkdirAll(filepath.Dir(dst), os.ModeDir|os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory for renaming file: %v", err)
		}
		if err := os.Rename(src, dst); err != nil {
			return fmt.Errorf("failed to move file: %v", err)
		}
	}

	if err := os.Remove(src); err != nil {
		return fmt.Errorf("failed to remove source: %v", err)
	}

	return nil
}

func moveNestedFolderIfExist(filePath string) error {
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
