package config

import (
	"buildenv/pkg/fileio"
	"fmt"
	"os"
	"path/filepath"
)

type CacheDir struct {
	Dir      string `json:"dir"`
	Readable bool   `json:"readable"`
	Writable bool   `json:"writable"`
}

func (c CacheDir) Read(archiveName, destDir string) (bool, error) {
	binaryPath := filepath.Join(c.Dir, archiveName)
	if !fileio.PathExists(binaryPath) {
		return false, nil // not an error even not exist.
	}

	if !fileio.IsReadable(binaryPath) {
		return false, fmt.Errorf("binary %s is not readable", binaryPath)
	}

	if err := os.MkdirAll(destDir, os.ModeDir|os.ModePerm); err != nil {
		return false, err
	}

	if err := fileio.Extract(binaryPath, destDir); err != nil {
		return false, err
	}

	return true, nil
}

func (c CacheDir) Write(packageDir string) error {
	if !c.Writable {
		return nil
	}

	if !fileio.PathExists(c.Dir) {
		return fmt.Errorf("cache dir %s does not exist", c.Dir)
	}
	if !fileio.IsWritable(c.Dir) {
		return fmt.Errorf("cache dir %s is not writable", c.Dir)
	}
	if !fileio.PathExists(packageDir) {
		return fmt.Errorf("package dir %s does not exist", packageDir)
	}

	// Create a tarball from package dir.
	archiveName := filepath.Base(packageDir) + ".tar.gz"
	destPath := filepath.Join(os.TempDir(), archiveName)
	if err := fileio.Targz(destPath, packageDir, false); err != nil {
		return err
	}

	// Remove the old tarball.
	if err := os.RemoveAll(filepath.Join(c.Dir, archiveName)); err != nil {
		return err
	}

	// Move the tarball to cache dir.
	if err := fileio.CopyFile(destPath, filepath.Join(c.Dir, archiveName)); err != nil {
		return err
	}

	defer os.Remove(destPath)
	return nil
}
