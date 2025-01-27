package config

import (
	"buildenv/pkg/fileio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type CacheDir struct {
	Dir      string `json:"dir"`
	Readable bool   `json:"readable"`
	Writable bool   `json:"writable"`
}

func (c CacheDir) Validate() error {
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

func (c CacheDir) Read(platformName, projectName, buildType, archiveName, destDir string) (bool, error) {
	archivePath := filepath.Join(c.Dir, platformName, projectName, buildType, archiveName)
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
	parts := strings.Split(filepath.Base(packageDir), "@")
	if len(parts) != 5 {
		return fmt.Errorf("invalid package dir: %s", packageDir)
	}

	var (
		libName      = parts[0]
		libVersion   = parts[1]
		platformName = parts[2]
		projectName  = parts[3]
		buildType    = parts[4]
	)

	archiveName := fmt.Sprintf("%s@%s.tar.gz", libName, libVersion)
	destPath := filepath.Join(os.TempDir(), archiveName)
	if err := fileio.Targz(destPath, packageDir, false); err != nil {
		return err
	}
	defer os.Remove(destPath)

	destDir := filepath.Join(c.Dir, platformName, projectName, buildType)

	// Remove the old tarball.
	if err := os.RemoveAll(filepath.Join(destDir, archiveName)); err != nil {
		return err
	}

	// Create the dir if not exist.
	if err := os.MkdirAll(destDir, os.ModeDir|os.ModePerm); err != nil {
		return err
	}

	// Move the tarball to cache dir.
	if err := fileio.CopyFile(destPath, filepath.Join(destDir, archiveName)); err != nil {
		return err
	}

	defer os.Remove(destPath)
	return nil
}
