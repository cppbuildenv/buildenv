package config

import (
	"buildenv/pkg/io"
	"fmt"
	"path/filepath"
	"strings"
)

type RootFS struct {
	Url     string    `json:"url"`
	RunPath string    `json:"run_path"`
	EnvVars RootFSEnv `json:"env_vars"`
}

type RootFSEnv struct {
	SYSROOT                string   `json:"SYSROOT"`
	PKG_CONFIG_SYSROOT_DIR string   `json:"PKG_CONFIG_SYSROOT_DIR"`
	PKG_CONFIG_PATH        []string `json:"PKG_CONFIG_PATH"`
}

func (r RootFS) Verify(args VerifyArgs) error {
	if r.Url == "" {
		return fmt.Errorf("rootfs.url is empty")
	}

	if r.RunPath == "" {
		return fmt.Errorf("rootfs.run_path is empty")
	}

	if r.EnvVars.SYSROOT == "" {
		return fmt.Errorf("rootfs.env.SYSROOT is empty")
	}

	if r.EnvVars.PKG_CONFIG_SYSROOT_DIR == "" {
		return fmt.Errorf("rootfs.env.PKG_CONFIG_SYSROOT_DIR is empty")
	}

	if len(r.EnvVars.PKG_CONFIG_PATH) == 0 {
		return fmt.Errorf("rootfs.env.PKG_CONFIG_PATH is empty")
	}

	if !args.CheckAndRepair {
		return nil
	}

	return r.checkAndRepair()
}

func (b RootFS) checkAndRepair() error {
	rootfsPath := filepath.Join(Dirs.DownloadDir, b.RunPath)
	if pathExists(rootfsPath) {
		return nil
	}

	// Download to fixed dir.
	downloaded, err := io.Download(b.Url, Dirs.DownloadDir)
	if err != nil {
		return fmt.Errorf("%s: download rootfs failed: %w", b.Url, err)
	}

	// Extract archive file.
	fileName := filepath.Base(b.Url)
	folderName := strings.TrimSuffix(fileName, ".tar.gz")
	extractPath := filepath.Join(Dirs.DownloadDir, folderName)
	if err := io.Extract(downloaded, extractPath); err != nil {
		return fmt.Errorf("%s: extract rootfs failed: %w", downloaded, err)
	}

	fmt.Printf("[✔] -------- %s (rootfs: %s).\n\n", filepath.Base(b.Url), extractPath)
	return nil
}
