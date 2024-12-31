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
	Url             string   `json:"url"`                    // Download url.
	ArchiveName     string   `json:"archive_name,omitempty"` // Archive name can be changed to avoid conflict.
	Path            string   `json:"path"`                   // Runtime path of tool, it's relative path  and would be converted to absolute path later.
	PkgConfigLibdir []string `json:"pkg_config_libdir"`

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

	// Add pkg-config libdir in rootfs to environment.
	var pkgConfigLibdirs []string
	for _, libdir := range r.PkgConfigLibdir {
		libDirFullPath := filepath.Join(r.fullpath, libdir)
		if !io.PathExists(libDirFullPath) {
			continue
		}

		pkgConfigLibdirs = append(pkgConfigLibdirs, libDirFullPath)
	}
	pkgConfigLibdirPaths := strings.Join(pkgConfigLibdirs, string(os.PathListSeparator))
	os.Setenv("PKG_CONFIG_LIBDIR", pkgConfigLibdirPaths)

	// TODO: this would make pkg-config cannot find libraries outside of rootfs.
	// os.Setenv("PKG_CONFIG_SYSROOT_DIR", r.fullpath)

	return nil
}

func (r RootFS) CheckAndRepair(request VerifyRequest) error {
	if !request.RepairBuildenv() {
		return nil
	}

	// Default folder name is the first folder name of archive name.
	// but it can be specified by archive name.
	folderName := strings.Split(r.Path, string(filepath.Separator))[0]
	if r.ArchiveName != "" {
		folderName = io.FileBaseName(r.ArchiveName)
	}
	location := filepath.Join(Dirs.ExtractedToolsDir, folderName)

	// Check if tool exists.
	if io.PathExists(r.fullpath) {
		if !request.Silent() {
			title := color.Sprintf(color.Green, "\n[✔] ---- Rootfs: %s\n", io.FileBaseName(r.Url))
			fmt.Printf("%sLocation: %s\n", title, location)
		}
		return nil
	}

	// Use archive name as download file name if specified.
	archiveName := filepath.Base(r.Url)
	if r.ArchiveName != "" {
		archiveName = r.ArchiveName
	}

	// Check and repair resource.
	repair := io.NewResourceRepair(r.Url, archiveName, folderName, Dirs.ExtractedToolsDir, Dirs.DownloadRootDir)
	if err := repair.CheckAndRepair(); err != nil {
		return err
	}

	// Print download & extract info.
	if !request.Silent() {
		title := color.Sprintf(color.Green, "\n[✔] ---- Rootfs: %s\n", io.FileBaseName(r.Url))
		fmt.Printf("%sLocation: %s\n", title, location)
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

	// Add pkg-config libdir in rootfs to environment.
	var pkgConfigLibdirs []string
	for _, libdir := range r.PkgConfigLibdir {
		if io.PathExists(filepath.Join(r.fullpath, libdir)) {
			pkgConfigLibdirs = append(pkgConfigLibdirs, filepath.Join(r.cmakepath, libdir))
		}
	}
	if len(pkgConfigLibdirs) > 0 {
		pkgConfigLibdirPaths := strings.Join(pkgConfigLibdirs, string(os.PathListSeparator))
		toolchain.WriteString(fmt.Sprintf("set(ENV{PKG_CONFIG_LIBDIR} \"%s\")\n", pkgConfigLibdirPaths))
	}

	// Write variables to environment
	environment.WriteString("\n# Set rootfs for cross compile.\n")
	environment.WriteString(fmt.Sprintf("export SYSROOT=%s\n", r.cmakepath))
	environment.WriteString(fmt.Sprintf("export PATH=%s\n", env.Join("${SYSROOT}", "${PATH}")))
	environment.WriteString("export PKG_CONFIG_SYSROOT_DIR=${SYSROOT}\n")

	return nil
}
