package config

import (
	"buildenv/pkg/io"
	"fmt"
	"net/url"
	"path/filepath"
)

type RootFS struct {
	Url     string    `json:"url"`
	Path    string    `json:"path"`
	EnvVars RootFSEnv `json:"env_vars"`
	None    bool      `json:"none"`
}

func (r RootFS) AbsolutePath() string {
	fullPath := filepath.Join(WorkspaceDir, r.Path)
	path, err := filepath.Abs(fullPath)
	if err != nil {
		panic(fmt.Sprintf("cannot get absolute path: %s", r.Path))
	}
	return path
}

type RootFSEnv struct {
	SYSROOT                string   `json:"SYSROOT"`
	PKG_CONFIG_SYSROOT_DIR string   `json:"PKG_CONFIG_SYSROOT_DIR"`
	PKG_CONFIG_PATH        []string `json:"PKG_CONFIG_PATH"`
}

func (r RootFS) Verify(host string, onlyFields bool) error {
	// If none is true, then rootfs is not required.
	if r.None {
		return nil
	}

	if r.Url == "" {
		return fmt.Errorf("rootfs.url is empty")
	}
	if r.Path == "" {
		return fmt.Errorf("rootfs.path is empty")
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

	if onlyFields {
		return nil
	}

	return r.checkIntegrity(host)
}

func (b RootFS) checkIntegrity(host string) error {
	rootfsPath := filepath.Join(WorkspaceDir, b.Path)
	if !pathExists(rootfsPath) {
		fullUrl, err := url.JoinPath(host, b.Url)
		if err != nil {
			return fmt.Errorf("buildenv.rootfs.url error: %w", err)
		}

		// Download to fixed dir.
		downloaded, err := io.Download(fullUrl, DownloadDir)
		if err != nil {
			return fmt.Errorf("%s: download rootfs failed: %w", fullUrl, err)
		}

		// Extract to dir with same parent.
		parentDir := filepath.Dir(b.Url)
		extractDir := filepath.Join(WorkspaceDir, parentDir)
		if err := io.Extract(downloaded, extractDir); err != nil {
			return fmt.Errorf("%s: extract rootfs failed: %w", downloaded, err)
		}

		fmt.Printf("[✔] ---- rootfs of platform setup success.\n\n")
	}
	return nil
}
