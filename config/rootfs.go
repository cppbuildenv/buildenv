package config

import (
	"buildenv/pkg/io"
	"fmt"
	"path/filepath"
)

type RootFS struct {
	Url         string    `json:"url"`
	ExtractPath string    `json:"extract_path"`
	RuntimePath string    `json:"runtime_path"`
	EnvVars     RootFSEnv `json:"env_vars"`
	None        bool      `json:"none"`
}

func (r RootFS) AbsolutePath() string {
	fullPath := filepath.Join(WorkspaceDir, r.RuntimePath)
	path, err := filepath.Abs(fullPath)
	if err != nil {
		panic(fmt.Sprintf("cannot get absolute path: %s", r.RuntimePath))
	}
	return path
}

type RootFSEnv struct {
	SYSROOT                string   `json:"SYSROOT"`
	PKG_CONFIG_SYSROOT_DIR string   `json:"PKG_CONFIG_SYSROOT_DIR"`
	PKG_CONFIG_PATH        []string `json:"PKG_CONFIG_PATH"`
}

func (r RootFS) Verify(checkAndRepiar bool) error {
	// If none is true, then rootfs is not required.
	if r.None {
		return nil
	}

	if r.Url == "" {
		return fmt.Errorf("rootfs.url is empty")
	}

	if r.ExtractPath == "" {
		return fmt.Errorf("rootfs.extract_path is empty")
	}

	if r.RuntimePath == "" {
		return fmt.Errorf("rootfs.runtime_path is empty")
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

	if !checkAndRepiar {
		return nil
	}

	return r.checkAndRepair()
}

func (b RootFS) checkAndRepair() error {
	rootfsPath := filepath.Join(WorkspaceDir, b.RuntimePath)
	if pathExists(rootfsPath) {
		return nil
	}

	// Download to fixed dir.
	downloaded, err := io.Download(b.Url, DownloadDir)
	if err != nil {
		return fmt.Errorf("%s: download rootfs failed: %w", b.Url, err)
	}

	// Extract to `extract_path`
	extractDir := filepath.Join(WorkspaceDir, b.ExtractPath)
	if err := io.Extract(downloaded, extractDir); err != nil {
		return fmt.Errorf("%s: extract rootfs failed: %w", downloaded, err)
	}

	fmt.Printf("[✔] ---- rootfs of platform setup success.\n\n")
	return nil
}
