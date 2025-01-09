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

func (c CacheDir) Verify() error {
	if c.Dir == "" {
		return fmt.Errorf("cache dir is empty")
	}
	if !fileio.PathExists(c.Dir) {
		return fmt.Errorf("cache dir %s does not exist", c.Dir)
	}
	if c.Readable && !fileio.IsReadable(c.Dir) {
		return fmt.Errorf("cache dir %s is not readable", c.Dir)
	}
	if c.Writable && !fileio.IsWritable(c.Dir) {
		return fmt.Errorf("cache dir %s is not writable", c.Dir)
	}
	return nil
}

func (c CacheDir) Read(archiveName, destDir string) (bool, error) {
	archivePath := filepath.Join(c.Dir, archiveName)
	if !fileio.PathExists(archivePath) {
		return false, nil // not an error even not exist.
	}

	if err := os.MkdirAll(destDir, os.ModeDir|os.ModePerm); err != nil {
		return false, err
	}

	if err := fileio.Extract(archivePath, destDir); err != nil {
		return false, err
	}

	return true, nil
}

func (c CacheDir) Write(packageDir string) error {
	if !c.Writable {
		return nil
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
