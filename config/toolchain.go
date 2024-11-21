package config

import (
	"buildenv/pkg/color"
	"buildenv/pkg/io"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Toolchain struct {
	Url             string `json:"url"`                    // Download url.
	ArchiveName     string `json:"archive_name,omitempty"` // Archive name can be changed to avoid conflict.
	Path            string `json:"path"`                   // Runtime path of tool, it's relative path  and would be converted to absolute path later.
	SystemName      string `json:"system_name"`            // System name, it will be used to generate toolchain file.
	SystemProcessor string `json:"system_processor"`       // System processor, it will be used to generate toolchain file.
	ToolchainPrefix string `json:"toolchain_prefix"`       // It'll be joined with toolchain path to generate toolchain file.
	CC              string `json:"cc"`
	CXX             string `json:"cxx"`
	FC              string `json:"fc"`
	RANLIB          string `json:"ranlib"`
	AR              string `json:"ar"`
	LD              string `json:"ld"`
	NM              string `json:"nm"`
	OBJDUMP         string `json:"objdump"`
	STRIP           string `json:"strip"`

	// Internal fields.
	fullpath string `json:"-"`
}

func (t *Toolchain) Verify(args VerifyArgs) error {
	// Verify toolchain download url.
	if t.Url == "" {
		return fmt.Errorf("toolchain.url is empty")
	}
	if err := io.CheckAvailable(t.Url); err != nil {
		return fmt.Errorf("toolchain.url is not accessible: %w", err)
	}

	// Verify toolchain path and convert to absolute path.
	if t.Path == "" {
		return fmt.Errorf("toolchain.path is empty")
	}
	toolchainPath, err := io.ToAbsPath(Dirs.DownloadRootDir, t.Path)
	if err != nil {
		return fmt.Errorf("cannot get absolute path: %s", t.Path)
	}
	t.fullpath = toolchainPath

	// This is used to cross-compile other ports by buildenv.
	os.Setenv("PATH", fmt.Sprintf("%s:%s", t.fullpath, os.Getenv("PATH")))

	if t.SystemName == "" {
		return fmt.Errorf("toolchain.system_name is empty")
	}

	if t.SystemProcessor == "" {
		return fmt.Errorf("toolchain.system_processor is empty")
	}

	// Verify toolchain prefix path and convert to absolute path.
	if t.ToolchainPrefix == "" {
		return fmt.Errorf("toolchain.toolchain_prefix is empty")
	}
	t.ToolchainPrefix = filepath.Join(t.fullpath, t.ToolchainPrefix)

	if t.CC == "" {
		return fmt.Errorf("toolchain.cc is empty")
	}

	if t.CXX == "" {
		return fmt.Errorf("toolchain.cxx is empty")
	}

	// This is used to cross-compile other ports by buildenv.
	os.Setenv("TOOLCHAIN_PREFIX", t.ToolchainPrefix)
	os.Setenv("CC", t.CC)
	os.Setenv("CXX", t.CXX)
	if t.FC != "" {
		os.Setenv("FC", t.FC)
	}
	if t.RANLIB != "" {
		os.Setenv("RANLIB", t.RANLIB)
	}
	if t.AR != "" {
		os.Setenv("AR", t.AR)
	}
	if t.LD != "" {
		os.Setenv("LD", t.LD)
	}
	if t.NM != "" {
		os.Setenv("NM", t.NM)
	}
	if t.OBJDUMP != "" {
		os.Setenv("OBJDUMP", t.OBJDUMP)
	}
	if t.STRIP != "" {
		os.Setenv("STRIP", t.STRIP)
	}

	if !args.CheckAndRepair() {
		return nil
	}

	return t.checkAndRepair()
}

func (t Toolchain) checkAndRepair() error {
	// Check if tool exists.
	if io.PathExists(t.fullpath) {
		return nil
	}

	// Use archive name as download file name if specified.
	archiveName := filepath.Base(t.Url)
	if t.ArchiveName != "" {
		archiveName = t.ArchiveName
	}

	// Check if need to download file.
	downloaded := filepath.Join(Dirs.DownloadRootDir, archiveName)
	if io.PathExists(downloaded) {
		// Redownload if remote file size and local file size not match.
		fileSize, err := io.FileSize(t.Url)
		if err != nil {
			return fmt.Errorf("%s: get remote filesize failed: %w", archiveName, err)
		}
		info, err := os.Stat(downloaded)
		if err != nil {
			return fmt.Errorf("%s: get local filesize failed: %w", archiveName, err)
		}
		if info.Size() != fileSize {
			if _, err := io.Download(t.Url, Dirs.DownloadRootDir, archiveName); err != nil {
				return fmt.Errorf("%s: download failed: %w", archiveName, err)
			}
		}
	} else {
		if _, err := io.Download(t.Url, Dirs.DownloadRootDir, archiveName); err != nil {
			return fmt.Errorf("%s: download failed: %w", archiveName, err)
		}
	}

	// Extract archive file.
	folderName := strings.Split(t.Path, string(filepath.Separator))[0]
	if t.ArchiveName != "" {
		folderName = io.FileBaseName(t.ArchiveName)
	}
	if err := io.Extract(downloaded, filepath.Join(Dirs.DownloadRootDir, folderName)); err != nil {
		return fmt.Errorf("%s: extract toolchain failed: %w", downloaded, err)
	}

	// Check if has nested folder (handling case where there's an extra nested folder).
	extractPath := filepath.Join(Dirs.DownloadRootDir, folderName)
	if err := io.MoveNestedFolderIfExist(extractPath); err != nil {
		return fmt.Errorf("%s: failed to move nested folder: %w", archiveName, err)
	}

	// Print download & extract info.
	fmt.Print(color.Sprintf(color.Blue, "[✔] -------- %s (toolchain: %s)\n\n", filepath.Base(t.Url), extractPath))
	return nil
}

func (t Toolchain) generate(toolchain, environment *strings.Builder) error {
	toolchain.WriteString("\n# Set toolchain for cross-compile.\n")
	toolchain.WriteString(fmt.Sprintf("set(ENV{PATH} \"%s:$ENV{PATH}\")\n", t.fullpath))

	writeIfNotEmpty := func(content, env, value string) {
		if value != "" {
			// Set toolchain variables.
			toolchain.WriteString(fmt.Sprintf("set(%s\"%s\")\n", content, value))

			// Set environment variables for makefile project.
			environment.WriteString(fmt.Sprintf("export %s=%s\n", env, value))
		}
	}

	environment.WriteString("# Set toolchain for cross compile.\n")
	environment.WriteString(fmt.Sprintf("export TOOLCHAIN_PATH=%s\n", t.fullpath))
	environment.WriteString("export PATH=${TOOLCHAIN_PATH}:${PATH}\n\n")

	writeIfNotEmpty("CMAKE_C_COMPILER 		", "CC", t.CC)
	writeIfNotEmpty("CMAKE_CXX_COMPILER		", "CXX", t.CXX)
	writeIfNotEmpty("CMAKE_Fortran_COMPILER	", "FC", t.FC)
	writeIfNotEmpty("CMAKE_RANLIB 			", "RANLIB", t.RANLIB)
	writeIfNotEmpty("CMAKE_AR 				", "AR", t.AR)
	writeIfNotEmpty("CMAKE_LINKER 			", "LD", t.LD)
	writeIfNotEmpty("CMAKE_NM 				", "NM", t.NM)
	writeIfNotEmpty("CMAKE_OBJDUMP 			", "OBJDUMP", t.OBJDUMP)
	writeIfNotEmpty("CMAKE_STRIP 			", "STRIP", t.STRIP)

	toolchain.WriteString("\n")
	toolchain.WriteString("set(CMAKE_C_FLAGS_INIT \"--sysroot=${CMAKE_SYSROOT}\")\n")
	toolchain.WriteString("set(CMAKE_CXX_FLAGS_INIT \"--sysroot=${CMAKE_SYSROOT}\")\n")

	return nil
}
