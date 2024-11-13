package config

import (
	"buildenv/pkg/color"
	"buildenv/pkg/io"
	"fmt"
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

func (t Toolchain) Verify(args VerifyArgs) error {
	if t.Url == "" {
		return fmt.Errorf("toolchain.url is empty")
	}

	if t.Path == "" {
		return fmt.Errorf("toolchain.path is empty")
	}

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
	toolchainPath := filepath.Join(Dirs.DownloadRootDir, t.Path)
	absToolchainPath, err := filepath.Abs(toolchainPath)
	if err != nil {
		return fmt.Errorf("cannot get absolute path of toolchain path: %s", toolchainPath)
	}

	toolchain.WriteString(fmt.Sprintf("set(ENV{PATH} \"%s:$ENV{PATH}\")\n", absToolchainPath))

	writeIfNotEmpty := func(content, env, value string) {
		if value != "" {
			// Set toolchain variables.
			toolchain.WriteString(fmt.Sprintf("set(%s\"%s\")\n", content, value))

			// Set environment variables for makefile project.
			environment.WriteString(fmt.Sprintf("export %s=%s\n", env, value))
		}
	}

	environment.WriteString("# Set toolchain for cross compile.\n")
	environment.WriteString(fmt.Sprintf("export TOOLCHAIN_PATH=%s\n", absToolchainPath))
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

	return nil
}

func (t Toolchain) checkAndRepair() error {
	toolchainPath := filepath.Join(Dirs.DownloadRootDir, t.Path)
	if pathExists(toolchainPath) {
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
