package config

import (
	"buildenv/pkg/color"
	"buildenv/pkg/env"
	"buildenv/pkg/io"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type RootFS struct {
	Url           string   `json:"url"`                    // Download url.
	ArchiveName   string   `json:"archive_name,omitempty"` // Archive name can be changed to avoid conflict.
	Path          string   `json:"path"`                   // Runtime path of tool, it's relative path  and would be converted to absolute path later.
	PkgConfigPath []string `json:"pkg_config_path"`        // Pkg config path, default will be `usr/lib/pkgconfig`

	// Internal fields.
	fullpath  string `json:"-"`
	cmakepath string `json:"-"`
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

	r.fullpath = filepath.Join(Dirs.ExtractedToolsDir, r.Path)
	r.cmakepath = fmt.Sprintf("${BUILDENV_ROOT_DIR}/downloads/tools/%s", r.Path)

	// This is for cross-compile other ports by buildenv.
	os.Setenv("SYSROOT", r.fullpath)
	os.Setenv("PKG_CONFIG_SYSROOT_DIR", r.fullpath)

	// Verify pkg_config_path and convert to absolute path.
	if len(r.PkgConfigPath) > 0 {
		var paths []string
		for _, itemPath := range r.PkgConfigPath {
			paths = append(paths, filepath.Join(r.fullpath, itemPath))
		}

		// This is for cross-compile other ports by buildenv.
		os.Setenv("PKG_CONFIG_PATH", strings.Join(paths, ":"))
	}

	if !args.CheckAndRepair() {
		return nil
	}

	return r.checkAndRepair()
}

func (r RootFS) checkAndRepair() error {
	// Check if tool exists.
	if io.PathExists(r.fullpath) {
		return nil
	}

	// Use archive name as download file name if specified.
	archiveName := filepath.Base(r.Url)
	if r.ArchiveName != "" {
		archiveName = r.ArchiveName
	}

	// Check if need to download file.
	downloaded := filepath.Join(Dirs.DownloadRootDir, archiveName)
	if io.PathExists(downloaded) {
		// Redownload if remote file size and local file size not match.
		fileSize, err := io.FileSize(r.Url)
		if err != nil {
			return fmt.Errorf("%s: get remote filesize failed: %w", archiveName, err)
		}
		info, err := os.Stat(downloaded)
		if err != nil {
			return fmt.Errorf("%s: get local filesize failed: %w", archiveName, err)
		}
		if info.Size() != fileSize {
			if _, err := io.Download(r.Url, Dirs.DownloadRootDir, archiveName); err != nil {
				return fmt.Errorf("%s: download failed: %w", archiveName, err)
			}
		}
	} else {
		if _, err := io.Download(r.Url, Dirs.DownloadRootDir, archiveName); err != nil {
			return fmt.Errorf("%s: download failed: %w", archiveName, err)
		}
	}

	// Extract archive file.
	folderName := strings.Split(r.Path, string(filepath.Separator))[0]
	if r.ArchiveName != "" {
		folderName = io.FileBaseName(r.ArchiveName)
	}
	if err := io.Extract(downloaded, filepath.Join(Dirs.ExtractedToolsDir, folderName)); err != nil {
		return fmt.Errorf("%s: extract rootfs failed: %w", downloaded, err)
	}

	// Check if has nested folder (handling case where there's an extra nested folder).
	extractPath := filepath.Join(Dirs.ExtractedToolsDir, folderName)
	if err := io.MoveNestedFolderIfExist(extractPath); err != nil {
		return fmt.Errorf("%s: failed to move nested folder: %w", archiveName, err)
	}

	// Print download & extract info.
	fmt.Print(color.Sprintf(color.Blue, "[✔] -------- %s (rootfs: %s)\n\n", filepath.Base(r.Url), extractPath))
	return nil
}

func (r RootFS) generate(toolchain, environment *strings.Builder) error {
	toolchain.WriteString("\n# Set sysroot for cross-compile.\n")
	toolchain.WriteString(fmt.Sprintf("set(CMAKE_SYSROOT \"%s\")\n", r.cmakepath))
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
		toolchain.WriteString(fmt.Sprintf("set(ENV{PKG_CONFIG_PATH} \"${CMAKE_SYSROOT}/%s%s$ENV{PKG_CONFIG_PATH}\")\n", path, string(os.PathListSeparator)))
	}

	// Write variables to buildenv.sh
	environment.WriteString("\n# Set rootfs for cross compile.\n")
	environment.WriteString(fmt.Sprintf("export SYSROOT=%s\n", r.cmakepath))
	environment.WriteString(fmt.Sprintf("export PATH=%s\n", env.Join("${SYSROOT}", "${PATH}")))
	environment.WriteString("export PKG_CONFIG_SYSROOT_DIR=${SYSROOT}\n")

	for index, path := range r.PkgConfigPath {
		fullpath := filepath.Join("${SYSROOT}", path)
		environment.WriteString(fmt.Sprintf("export PKG_CONFIG_PATH=%s\n", env.Join(fullpath, "${PKG_CONFIG_PATH}")))

		if index == len(r.PkgConfigPath)-1 {
			environment.WriteString("\n")
		}
	}

	// Set the environment variables.
	os.Setenv("SYSROOT", r.fullpath)
	os.Setenv("PKG_CONFIG_SYSROOT_DIR", r.fullpath)
	os.Setenv("PKG_CONFIG_PATH", strings.Join(r.PkgConfigPath, ":"))
	os.Setenv("PATH", fmt.Sprintf("%s%c%s", r.fullpath, os.PathListSeparator, os.Getenv("PATH")))

	return nil
}
