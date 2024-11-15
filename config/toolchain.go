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
	Url             string          `json:"url"`
	Path            string          `json:"path"`
	SystemName      string          `json:"system_name"`
	SystemProcessor string          `json:"system_processor"`
	EnvVars         ToolchainEnvVar `json:"env_vars"`
}

type ToolchainEnvVar struct {
	CC      string `json:"CC"`
	CXX     string `json:"CXX"`
	FC      string `json:"FC"`
	RANLIB  string `json:"RANLIB"`
	AR      string `json:"AR"`
	LD      string `json:"LD"`
	NM      string `json:"NM"`
	OBJDUMP string `json:"OBJDUMP"`
	STRIP   string `json:"STRIP"`
}

func (t *Toolchain) Verify(args VerifyArgs) error {
	// Relative path -> Absolute path.
	var toAbsPath = func(relativePath string) (string, error) {
		path := filepath.Join(Dirs.DownloadRootDir, relativePath)
		rootfsPath, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}

		return rootfsPath, nil
	}

	if t.Url == "" {
		return fmt.Errorf("toolchain.url is empty")
	}

	// Verify toolchain path and convert to absolute path.
	if t.Path == "" {
		return fmt.Errorf("toolchain.path is empty")
	}
	toolchainPath, err := toAbsPath(t.Path)
	if err != nil {
		return fmt.Errorf("cannot get absolute path: %s", t.Path)
	}
	t.Path = toolchainPath

	// This is used to cross-compile other ports by buildenv.
	os.Setenv("PATH", fmt.Sprintf("%s:%s", t.Path, os.Getenv("PATH")))

	if t.SystemName == "" {
		return fmt.Errorf("toolchain.system_name is empty")
	}

	if t.SystemProcessor == "" {
		return fmt.Errorf("toolchain.system_processor is empty")
	}

	if t.EnvVars.CC == "" {
		return fmt.Errorf("toolchain.env.CC is empty")
	}

	if t.EnvVars.CXX == "" {
		return fmt.Errorf("toolchain.env.CXX is empty")
	}

	if !args.CheckAndRepair {
		return nil
	}

	return t.checkAndRepair()
}

func (t Toolchain) generate(toolchain, environment *strings.Builder) error {
	toolchain.WriteString("\n# Set toolchain for cross-compile.\n")
	toolchain.WriteString(fmt.Sprintf("set(ENV{PATH} \"%s:$ENV{PATH}\")\n", t.Path))

	writeIfNotEmpty := func(content, env, value string) {
		if value != "" {
			// Set toolchain variables.
			toolchain.WriteString(fmt.Sprintf("set(%s\"%s\")\n", content, value))

			// Set environment variables for makefile project.
			environment.WriteString(fmt.Sprintf("export %s=%s\n", env, value))
		}
	}

	environment.WriteString("# Set toolchain for cross compile.\n")
	environment.WriteString(fmt.Sprintf("export TOOLCHAIN_PATH=%s\n", t.Path))
	environment.WriteString("export PATH=${TOOLCHAIN_PATH}:${PATH}\n\n")

	writeIfNotEmpty("CMAKE_C_COMPILER 		", "CC", t.EnvVars.CC)
	writeIfNotEmpty("CMAKE_CXX_COMPILER		", "CXX", t.EnvVars.CXX)
	writeIfNotEmpty("CMAKE_Fortran_COMPILER	", "FC", t.EnvVars.FC)
	writeIfNotEmpty("CMAKE_RANLIB 			", "RANLIB", t.EnvVars.RANLIB)
	writeIfNotEmpty("CMAKE_AR 				", "AR", t.EnvVars.AR)
	writeIfNotEmpty("CMAKE_LINKER 			", "LD", t.EnvVars.LD)
	writeIfNotEmpty("CMAKE_NM 				", "NM", t.EnvVars.NM)
	writeIfNotEmpty("CMAKE_OBJDUMP 			", "OBJDUMP", t.EnvVars.OBJDUMP)
	writeIfNotEmpty("CMAKE_STRIP 			", "STRIP", t.EnvVars.STRIP)

	toolchain.WriteString("\n")
	toolchain.WriteString("set(CMAKE_C_FLAGS_INIT \"--sysroot=${CMAKE_SYSROOT}\")\n")
	toolchain.WriteString("set(CMAKE_CXX_FLAGS_INIT \"--sysroot=${CMAKE_SYSROOT}\")\n")

	return nil
}

func (t Toolchain) checkAndRepair() error {
	// Check if tool exists.
	if io.PathExists(t.Path) {
		return nil
	}

	// Download to fixed dir.
	downloaded, err := io.Download(t.Url, Dirs.DownloadRootDir)
	if err != nil {
		return fmt.Errorf("%s: download toolchain failed: %w", t.Url, err)
	}

	// Extract archive file.
	fileName := filepath.Base(t.Url)
	folderName := strings.TrimSuffix(fileName, ".tar.gz")
	extractPath := filepath.Join(Dirs.DownloadRootDir, folderName)
	if err := io.Extract(downloaded, extractPath); err != nil {
		return fmt.Errorf("%s: extract toolchain failed: %w", downloaded, err)
	}

	fmt.Print(color.Sprintf(color.Blue, "[✔] -------- %s (toolchain: %s)\n\n", filepath.Base(t.Url), extractPath))
	return nil
}
