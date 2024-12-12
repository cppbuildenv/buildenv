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
	Url         string `json:"url"`                    // Download url.
	ArchiveName string `json:"archive_name,omitempty"` // Archive name can be changed to avoid conflict.
	Path        string `json:"path"`                   // Runtime path of tool, it's relative path  and would be converted to absolute path later.

	// Internal fields.
	fullpath  string `json:"-"`
	cmakepath string `json:"-"`
}

func (r *RootFS) Verify() error {
	// Verify rootfs download url.
	if r.Url == "" {
		return fmt.Errorf("rootfs.url is empty")
	}
	if err := io.CheckAvailable(r.Url); err != nil {
		return fmt.Errorf("rootfs.url of %s is not accessible", r.Url)
	}

	// Verify rootfs path and convert to absolute path.
	if r.Path == "" {
		return fmt.Errorf("rootfs.path is empty")
	}

	r.fullpath = filepath.Join(Dirs.ExtractedToolsDir, r.Path)
	r.cmakepath = fmt.Sprintf("${BUILDENV_ROOT_DIR}/downloads/tools/%s", r.Path)

	os.Setenv("SYSROOT", r.fullpath)
	return nil
}

func (r RootFS) CheckAndRepair(args VerifyArgs) error {
	if !args.CheckAndRepair() {
		return nil
	}

	// Default folder name is the first folder name of archive name.
	// but it can be specified by archive name.
	folderName := strings.Split(r.Path, string(filepath.Separator))[0]
	if r.ArchiveName != "" {
		folderName = io.FileBaseName(r.ArchiveName)
	}
	extractedPath := filepath.Join(Dirs.ExtractedToolsDir, folderName)

	// Check if tool exists.
	if io.PathExists(r.fullpath) {
		// No need to show rootfs state info when install a port.
		if args.PortToInstall() == "" && !args.Silent() {
			title := color.Sprintf(color.Green, "\n[✔] ---- Rootfs: %s\n", io.FileBaseName(r.Url))
			fmt.Printf("%sLocation: %s\n", title, extractedPath)
		}
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
			downloadRequest := io.NewDownloadRequest(r.Url, Dirs.DownloadRootDir)
			downloadRequest.SetArchiveName(archiveName)
			if _, err := downloadRequest.Download(); err != nil {
				return fmt.Errorf("%s: download failed: %w", archiveName, err)
			}
		}
	} else {
		downloadRequest := io.NewDownloadRequest(r.Url, Dirs.DownloadRootDir)
		downloadRequest.SetArchiveName(archiveName)
		if _, err := downloadRequest.Download(); err != nil {
			return fmt.Errorf("%s: download failed: %w", archiveName, err)
		}
	}

	// Extract archive file.
	if err := io.Extract(downloaded, filepath.Join(Dirs.ExtractedToolsDir, folderName)); err != nil {
		return fmt.Errorf("%s: extract rootfs failed: %w", downloaded, err)
	}

	// Check if has nested folder (handling case where there's an extra nested folder).
	if err := io.MoveNestedFolderIfExist(extractedPath); err != nil {
		return fmt.Errorf("%s: failed to move nested folder: %w", archiveName, err)
	}

	// Print download & extract info.
	if !args.Silent() {
		title := color.Sprintf(color.Green, "\n[✔] ---- Rootfs: %s\n", io.FileBaseName(r.Url))
		fmt.Printf("%sLocation: %s\n", title, extractedPath)
	}
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

	// Write variables to buildenv.sh
	environment.WriteString("\n# Set rootfs for cross compile.\n")
	environment.WriteString(fmt.Sprintf("export SYSROOT=%s\n", r.cmakepath))
	environment.WriteString(fmt.Sprintf("export PATH=%s\n", env.Join("${SYSROOT}", "${PATH}")))
	environment.WriteString("export PKG_CONFIG_SYSROOT_DIR=${SYSROOT}\n")

	return nil
}
