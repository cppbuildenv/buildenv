package config

import (
	"buildenv/pkg/color"
	"buildenv/pkg/io"
	"fmt"
	"os"
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
	rootfsPath := filepath.Join(Dirs.DownloadRootDir, b.RunPath)
	if pathExists(rootfsPath) {
		return nil
	}

	// Download to fixed dir.
	downloaded, err := io.Download(b.Url, Dirs.DownloadRootDir)
	if err != nil {
		return fmt.Errorf("%s: download rootfs failed: %w", b.Url, err)
	}

	// Extract archive file.
	fileName := filepath.Base(b.Url)
	folderName := strings.TrimSuffix(fileName, ".tar.gz")
	extractPath := filepath.Join(Dirs.DownloadRootDir, folderName)
	if err := io.Extract(downloaded, extractPath); err != nil {
		return fmt.Errorf("%s: extract rootfs failed: %w", downloaded, err)
	}

	fmt.Print(color.Sprintf(color.Blue, "[✔] -------- %s (rootfs: %s)\n\n", filepath.Base(b.Url), extractPath))
	return nil
}

func (r RootFS) generate(toolchain, environment *strings.Builder) error {
	rootfsPath := filepath.Join(Dirs.DownloadRootDir, r.RunPath)
	absRootFSPath, err := filepath.Abs(rootfsPath)
	if err != nil {
		panic(fmt.Sprintf("cannot get absolute path: %s", rootfsPath))
	}

	toolchain.WriteString("\n# Set sysroot for cross-compile.\n")
	toolchain.WriteString(fmt.Sprintf("set(CMAKE_SYSROOT \"%s\")\n", absRootFSPath))
	toolchain.WriteString("list(APPEND CMAKE_FIND_ROOT_PATH \"${CMAKE_SYSROOT}\")\n")
	toolchain.WriteString("list(APPEND CMAKE_PREFIX_PATH \"${CMAKE_SYSROOT}\")\n")

	// Search programs in the host environment.
	toolchain.WriteString("\n# Search programs in the host environment.\n")
	toolchain.WriteString("set(CMAKE_FIND_ROOT_PATH_MODE_PROGRAM NEVER)\n")

	// Search libraries and headers in the target environment.
	toolchain.WriteString("\n# Search libraries and headers in the target environment.\n")
	toolchain.WriteString("set(CMAKE_FIND_ROOT_PATH_MODE_LIBRARY ONLY)\n")
	toolchain.WriteString("set(CMAKE_FIND_ROOT_PATH_MODE_INCLUDE ONLY)\n")
	toolchain.WriteString("set(CMAKE_FIND_ROOT_PATH_MODE_PACKAGE ONLY)\n")

	toolchain.WriteString("\n# Set pkg-config path for cross-compile.\n")
	toolchain.WriteString("set(ENV{PKG_CONFIG_SYSROOT_DIR} \"${CMAKE_SYSROOT}\")\n")

	// Replace the path with the workspace directory.
	for i, path := range r.EnvVars.PKG_CONFIG_PATH {
		fullPath := filepath.Join(Dirs.DownloadRootDir, path)
		absPath, err := filepath.Abs(fullPath)
		if err != nil {
			return fmt.Errorf("cannot get absolute path: %s", fullPath)
		}

		r.EnvVars.PKG_CONFIG_PATH[i] = absPath
	}
	toolchain.WriteString(fmt.Sprintf("set(ENV{PKG_CONFIG_PATH} \"%s\")\n", strings.Join(r.EnvVars.PKG_CONFIG_PATH, ":")))

	// Set environment variables for makefile project.
	environment.WriteString("\n# Set rootfs for cross compile.\n")
	environment.WriteString(fmt.Sprintf("export SYSROOT=%s\n", absRootFSPath))
	environment.WriteString("export PATH=${SYSROOT}:${PATH}\n")
	environment.WriteString("export PKG_CONFIG_SYSROOT_DIR=${SYSROOT}\n")
	environment.WriteString(fmt.Sprintf("export PKG_CONFIG_PATH=%s\n\n", strings.Join(r.EnvVars.PKG_CONFIG_PATH, ":")))

	// Make sure the toolchain is in the PATH of current process.
	os.Setenv("SYSROOT", absRootFSPath)
	os.Setenv("PKG_CONFIG_SYSROOT_DIR", absRootFSPath)
	os.Setenv("PKG_CONFIG_PATH", strings.Join(r.EnvVars.PKG_CONFIG_PATH, ":"))
	os.Setenv("PATH", fmt.Sprintf("%s%c%s", absRootFSPath, os.PathListSeparator, os.Getenv("PATH")))

	return nil
}
