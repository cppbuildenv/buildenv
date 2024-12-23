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

type Toolchain struct {
	Url             string `json:"url"`                    // Download url or local file url.
	ArchiveName     string `json:"archive_name,omitempty"` // Archive name can be changed to avoid conflict.
	Path            string `json:"path"`                   // Runtime path of tool, it's relative path and would be converted to absolute path later.
	SystemName      string `json:"system_name"`            // It would be "Windows", "Linux", "Android" and so on.
	SystemProcessor string `json:"system_processor"`       // It would be "x86_64", "aarch64" and so on.
	Host            string `json:"host"`                   // It would be "x86_64-linux-gnu", "aarch64-linux-gnu" and so on.
	ToolchainPrefix string `json:"toolchain_prefix"`       // It would be like "x86_64-linux-gnu-"
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
	fullpath  string `json:"-"`
	cmakepath string `json:"-"`
}

func (t *Toolchain) Verify() error {
	rootfs := os.Getenv("SYSROOT")

	// Verify toolchain download url.
	if t.Url == "" {
		return fmt.Errorf("toolchain.url would be http url or local file url, but it's empty")
	}
	if err := io.CheckAvailable(t.Url); err != nil {
		return fmt.Errorf("toolchain.url of %s is not accessible", t.Url)
	}

	// Verify toolchain path and convert to absolute path.
	if t.Path == "" {
		return fmt.Errorf("toolchain.path is empty")
	}
	t.fullpath = filepath.Join(Dirs.ExtractedToolsDir, t.Path)
	t.cmakepath = fmt.Sprintf("${BUILDENV_ROOT_DIR}/downloads/tools/%s", t.Path)
	os.Setenv("PATH", t.fullpath+string(os.PathListSeparator)+os.Getenv("PATH"))

	if t.SystemName == "" {
		return fmt.Errorf("toolchain.system_name is empty")
	}

	if t.SystemProcessor == "" {
		return fmt.Errorf("toolchain.system_processor is empty")
	}

	// Verify toolchain prefix path and convert to absolute path.
	if t.ToolchainPrefix == "" {
		return fmt.Errorf("toolchain.toolchain_prefix should be like 'x86_64-linux-gnu-', but it's empty")
	}
	os.Setenv("TOOLCHAIN_PREFIX", t.ToolchainPrefix)

	if t.Host == "" {
		return fmt.Errorf("toolchain.host should be like 'x86_64-linux-gnu', but it's empty")
	}
	os.Setenv("HOST", t.Host)

	if t.CC == "" {
		return fmt.Errorf("toolchain.cc is empty")
	}
	os.Setenv("CC", fmt.Sprintf("%s --sysroot=%s", t.CC, rootfs))

	if t.CXX == "" {
		return fmt.Errorf("toolchain.cxx is empty")
	}
	os.Setenv("CXX", fmt.Sprintf("%s --sysroot=%s", t.CXX, rootfs))

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
		os.Setenv("CXX", fmt.Sprintf("%s --sysroot=%s", t.LD, rootfs))
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

	return nil
}

func (t Toolchain) CheckAndRepair(args VerifyArgs) error {
	if !args.CheckAndRepair() {
		return nil
	}

	// Default folder name is the first folder name of archive name.
	// but it can be specified by archive name.
	folderName := strings.Split(t.Path, string(filepath.Separator))[0]
	if t.ArchiveName != "" {
		folderName = io.FileBaseName(t.ArchiveName)
	}
	extractedPath := filepath.Join(Dirs.ExtractedToolsDir, folderName)

	// Check if tool exists.
	if io.PathExists(t.fullpath) {
		if !args.Silent() {
			title := color.Sprintf(color.Green, "\n[✔] ---- Toolchain: %s\n", io.FileBaseName(t.Url))
			fmt.Printf("%sLocation: %s\n", title, extractedPath)
		}

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
			downloadRequest := io.NewDownloadRequest(t.Url, Dirs.DownloadRootDir)
			downloadRequest.SetArchiveName(archiveName)
			if _, err := downloadRequest.Download(); err != nil {
				return fmt.Errorf("%s: download failed: %w", archiveName, err)
			}
		}
	} else {
		downloadRequest := io.NewDownloadRequest(t.Url, Dirs.DownloadRootDir)
		downloadRequest.SetArchiveName(archiveName)
		if _, err := downloadRequest.Download(); err != nil {
			return fmt.Errorf("%s: download failed: %w", archiveName, err)
		}
	}

	// Extract archive file.
	if err := io.Extract(downloaded, filepath.Join(Dirs.ExtractedToolsDir, folderName)); err != nil {
		return fmt.Errorf("%s: extract toolchain failed: %w", downloaded, err)
	}

	// Check if has nested folder (handling case where there's an extra nested folder).
	if err := io.MoveNestedFolderIfExist(extractedPath); err != nil {
		return fmt.Errorf("%s: failed to move nested folder: %w", archiveName, err)
	}

	// Print download & extract info.
	if !args.Silent() {
		title := color.Sprintf(color.Green, "\n[✔] ---- Toolchain: %s\n", io.FileBaseName(t.Url))
		fmt.Printf("%sLocation: %s\n", title, extractedPath)
	}
	return nil
}

func (t Toolchain) generate(toolchain, environment *strings.Builder) error {
	toolchain.WriteString("\n# Set toolchain for cross-compile.\n")
	toolchain.WriteString(fmt.Sprintf("set(ENV{PATH} \"%s\")\n", env.Join(t.cmakepath, "$ENV{PATH}")))

	writeIfNotEmpty := func(content, env, value string) {
		if value != "" {
			// Set toolchain variables.
			toolchain.WriteString(fmt.Sprintf("set(%s\"%s\")\n", content, value))

			// Set environment variables for makefile project.
			environment.WriteString(fmt.Sprintf("export %s=%s\n", env, value))
		}
	}

	environment.WriteString("\n# Set toolchain for cross compile.\n")
	environment.WriteString(fmt.Sprintf("export TOOLCHAIN_PATH=%s\n", t.cmakepath))
	environment.WriteString(fmt.Sprintf("export PATH=%s\n\n", env.Join("${TOOLCHAIN_PATH}", "${PATH}")))

	environment.WriteString("# Set cross compile tools.\n")
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
