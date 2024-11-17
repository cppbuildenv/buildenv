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
	Url           string   `json:"url"`
	Path          string   `json:"path"`
	PkgConfigPath []string `json:"pkg_config_path"`

	// Internal fields.
	fullpath string `json:"-"`
}

type RootFSEnv struct {
	SYSROOT                string   `json:"SYSROOT"`
	PKG_CONFIG_SYSROOT_DIR string   `json:"PKG_CONFIG_SYSROOT_DIR"`
	PKG_CONFIG_PATH        []string `json:"PKG_CONFIG_PATH"`
}

func (r *RootFS) Verify(args VerifyArgs) error {
	// Verify rootfs download url.
	if r.Url == "" {
		return fmt.Errorf("rootfs.url is empty")
	}
	if err := io.CheckAvailable(r.Url); err != nil {
		return fmt.Errorf("rootfs.url is not accessible: %w", err)
	}

	// Verify rootfs path and convert to absolute path.
	if r.Path == "" {
		return fmt.Errorf("rootfs.path is empty")
	}
	rootfsPath, err := io.ToAbsPath(Dirs.DownloadRootDir, r.Path)
	if err != nil {
		return fmt.Errorf("cannot get absolute path: %s", r.Path)
	}
	r.fullpath = rootfsPath

	// This is for cross-compile other ports by buildenv.
	os.Setenv("SYSROOT", rootfsPath)
	os.Setenv("PKG_CONFIG_SYSROOT_DIR", rootfsPath)

	// Verify pkg_config_path and convert to absolute path.
	if len(r.PkgConfigPath) > 0 {
		var paths []string
		for _, itemPath := range r.PkgConfigPath {
			absPath, err := io.ToAbsPath(Dirs.DownloadRootDir, filepath.Join(r.fullpath, itemPath))
			if err != nil {
				return fmt.Errorf("cannot get absolute path: %s", itemPath)
			}

			paths = append(paths, absPath)
		}

		// This is for cross-compile other ports by buildenv.
		os.Setenv("PKG_CONFIG_PATH", strings.Join(paths, ":"))
	}

	if !args.CheckAndRepair {
		return nil
	}

	return r.checkAndRepair()
}

func (r RootFS) checkAndRepair() error {
	// Check if tool exists.
	if io.PathExists(r.fullpath) {
		return nil
	}

	// Download to fixed dir.
	downloaded, err := io.Download(r.Url, Dirs.DownloadRootDir)
	if err != nil {
		return fmt.Errorf("%s: download rootfs failed: %w", r.Url, err)
	}

	// Extract archive file.
	fileName := filepath.Base(r.Url)
	folderName := strings.Split(r.Path, string(filepath.Separator))[0]
	if err := io.Extract(downloaded, filepath.Join(Dirs.DownloadRootDir, folderName)); err != nil {
		return fmt.Errorf("%s: extract rootfs failed: %w", downloaded, err)
	}

	// Check if has nested folder (handling case where there's an extra nested folder).
	extractPath := filepath.Join(Dirs.DownloadRootDir, folderName)
	if err := io.MoveNestedFolderIfExist(extractPath); err != nil {
		return fmt.Errorf("%s: failed to move nested folder: %w", fileName, err)
	}

	// Print download & extract info.
	fmt.Print(color.Sprintf(color.Blue, "[✔] -------- %s (rootfs: %s)\n\n", filepath.Base(r.Url), extractPath))
	return nil
}

func (r RootFS) generate(toolchain, environment *strings.Builder) error {
	toolchain.WriteString("\n# Set sysroot for cross-compile.\n")
	toolchain.WriteString(fmt.Sprintf("set(CMAKE_SYSROOT \"%s\")\n", r.fullpath))
	toolchain.WriteString("list(APPEND CMAKE_FIND_ROOT_PATH \"${CMAKE_SYSROOT}\")\n")

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
	for _, path := range r.PkgConfigPath {
		toolchain.WriteString(fmt.Sprintf("set(ENV{PKG_CONFIG_PATH} \"${CMAKE_SYSROOT}/%s:$ENV{PKG_CONFIG_PATH}\")\n", path))
	}

	// Write variables to buildenv.sh
	environment.WriteString("\n# Set rootfs for cross compile.\n")
	environment.WriteString(fmt.Sprintf("export SYSROOT=%s\n", r.fullpath))
	environment.WriteString("export PATH=${SYSROOT}:${PATH}\n")
	environment.WriteString("export PKG_CONFIG_SYSROOT_DIR=${SYSROOT}\n")
	environment.WriteString(fmt.Sprintf("export PKG_CONFIG_PATH=%s:$PKG_CONFIG_PATH\n", strings.Join(r.PkgConfigPath, ":")))

	// Set the environment variables.
	os.Setenv("SYSROOT", r.fullpath)
	os.Setenv("PKG_CONFIG_SYSROOT_DIR", r.fullpath)
	os.Setenv("PKG_CONFIG_PATH", strings.Join(r.PkgConfigPath, ":"))
	os.Setenv("PATH", fmt.Sprintf("%s%c%s", r.fullpath, os.PathListSeparator, os.Getenv("PATH")))

	return nil
}
