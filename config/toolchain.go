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
	if err := fileio.CheckAvailable(t.Url); err != nil {
		return fmt.Errorf("toolchain.url of %s is not accessible", t.Url)
	}

	switch {
	// Web resource file would be extracted to specified path, so path cannot be empty.
	case strings.HasPrefix(t.Url, "http"), strings.HasPrefix(t.Url, "ftp"):
		if t.Path == "" {
			return fmt.Errorf("toolchain.path is empty")
		}

		t.fullpath = filepath.Join(Dirs.ExtractedToolsDir, t.Path)
		t.cmakepath = fmt.Sprintf("${BUILDENV_ROOT_DIR}/downloads/tools/%s", t.Path)
		os.Setenv("PATH", t.fullpath+string(os.PathListSeparator)+os.Getenv("PATH"))

	case strings.HasPrefix(t.Url, "file:///"):
		localPath := strings.TrimPrefix(t.Url, "file:///")
		state, err := os.Stat(localPath)
		if err != nil {
			return fmt.Errorf("toolchain.url of %s is not accessible", t.Url)
		}

		if state.IsDir() {
			t.fullpath = filepath.Join(localPath, t.Path)
			t.cmakepath = t.fullpath
			os.Setenv("PATH", t.fullpath+string(os.PathListSeparator)+os.Getenv("PATH"))
		} else {
			// Even local must be a archive file and path should not be empty.
			if t.Path == "" {
				return fmt.Errorf("toolchain.path is empty")
			}

			// Check if buildenv supported archive file.
			if !fileio.IsSupportedArchive(localPath) {
				return fmt.Errorf("toolchain.path of %s is not a archive file", t.Url)
			}

			t.fullpath = filepath.Join(Dirs.ExtractedToolsDir, t.Path)
			t.cmakepath = fmt.Sprintf("${BUILDENV_ROOT_DIR}/downloads/tools/%s", t.Path)
			os.Setenv("PATH", t.fullpath+string(os.PathListSeparator)+os.Getenv("PATH"))
		}

	default:
		return fmt.Errorf("toolchain.url of %s is not accessible", t.Url)
	}

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

func (t Toolchain) CheckAndRepair(request VerifyRequest) error {
	if !request.RepairBuildenv() {
		return nil
	}

	// Default folder name is the first folder name of archive name.
	// but it can be specified by archive name.
	folderName := strings.Split(t.Path, string(filepath.Separator))[0]
	if t.ArchiveName != "" {
		folderName = fileio.FileBaseName(t.ArchiveName)
	}
	location := filepath.Join(Dirs.ExtractedToolsDir, folderName)

	// Check if tool exists.
	if fileio.PathExists(t.fullpath) {
		if !request.Silent() {
			title := color.Sprintf(color.Green, "\n[✔] ---- Toolchain: %s\n", fileio.FileBaseName(t.Url))
			fmt.Printf("%sLocation: %s\n", title, location)
		}

		return nil
	}

	// Use archive name as download file name if specified.
	archiveName := filepath.Base(t.Url)
	if t.ArchiveName != "" {
		archiveName = t.ArchiveName
	}

	// Check and repair resource.
	repair := fileio.NewResourceRepair(t.Url, archiveName, folderName, Dirs.ExtractedToolsDir, Dirs.DownloadRootDir)
	if err := repair.CheckAndRepair(); err != nil {
		return err
	}

	// Print download & extract info.
	if !request.Silent() {
		title := color.Sprintf(color.Green, "\n[✔] ---- Toolchain: %s\n", fileio.FileBaseName(t.Url))
		fmt.Printf("%sLocation: %s\n", title, location)
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
