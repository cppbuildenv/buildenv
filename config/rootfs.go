package config

import (
	"buildenv/pkg/color"
	"buildenv/pkg/env"
	"buildenv/pkg/fileio"
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

func (r *RootFS) Validate() error {
	// Validate rootfs download url.
	if r.Url == "" {
		return fmt.Errorf("rootfs.url is empty")
	}

	// Validate rootfs path and convert to absolute path.
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
		if !fileio.PathExists(libDirFullPath) {
			continue
		}

		pkgConfigLibdirs = append(pkgConfigLibdirs, libDirFullPath)
	}
	pkgConfigLibdirPaths := strings.Join(pkgConfigLibdirs, string(os.PathListSeparator))
	os.Setenv("PKG_CONFIG_LIBDIR", pkgConfigLibdirPaths)

	return nil
}

func (r RootFS) CheckAndRepair(args SetupArgs) error {
	if !args.RepairBuildenv() {
		return nil
	}

	// Default folder name is the first folder name of archive name.
	// but it can be specified by archive name.
	folderName := strings.Split(r.Path, string(filepath.Separator))[0]
	if r.ArchiveName != "" {
		folderName = fileio.FileBaseName(r.ArchiveName)
	}
	location := filepath.Join(Dirs.ExtractedToolsDir, folderName)

	// Check if tool exists.
	if fileio.PathExists(r.fullpath) {
		if !args.Silent() {
			title := color.Sprintf(color.Green, "\n[✔] ---- Rootfs: %s\n", fileio.FileBaseName(r.Url))
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
	repair := fileio.NewDownloadRepair(r.Url, archiveName, folderName, Dirs.ExtractedToolsDir, Dirs.DownloadedDir)
	if err := repair.CheckAndRepair(); err != nil {
		return err
	}

	// Print download & extract info.
	if !args.Silent() {
		title := color.Sprintf(color.Green, "\n[✔] ---- Rootfs: %s\n", fileio.FileBaseName(r.Url))
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
		if fileio.PathExists(filepath.Join(r.fullpath, libdir)) {
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
