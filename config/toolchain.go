package config

import (
	"buildenv/pkg/io"
	"fmt"
	"path/filepath"
	"strings"
)

type Toolchain struct {
	Url       string          `json:"url"`
	RunPath   string          `json:"run_path"`
	EnvVars   ToolchainEnvVar `json:"env_vars"`
	CMakeVars ToolChainVars   `json:"cmake_vars"`
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

type ToolChainVars struct {
	CMAKE_SYSTEM_NAME      string `json:"CMAKE_SYSTEM_NAME"`
	CMAKE_SYSTEM_PROCESSOR string `json:"CMAKE_SYSTEM_PROCESSOR"`
}

func (t Toolchain) Verify(checkAndRepiar bool) error {
	if t.Url == "" {
		return fmt.Errorf("toolchain.url is empty")
	}

	if t.RunPath == "" {
		return fmt.Errorf("toolchain.run_path is empty")
	}

	if t.EnvVars.CC == "" {
		return fmt.Errorf("toolchain.env.CC is empty")
	}

	if t.EnvVars.CXX == "" {
		return fmt.Errorf("toolchain.env.CXX is empty")
	}

	if t.CMakeVars.CMAKE_SYSTEM_NAME == "" {
		return fmt.Errorf("toolchain.cmake_vars.CMAKE_SYSTEM_NAME is empty")
	}

	if t.CMakeVars.CMAKE_SYSTEM_PROCESSOR == "" {
		return fmt.Errorf("toolchain.cmake_vars.CMAKE_SYSTEM_PROCESSOR is empty")
	}

	if !checkAndRepiar {
		return nil
	}

	return t.checkAndRepair()
}

func (t Toolchain) checkAndRepair() error {
	toolchainPath := filepath.Join(WorkspaceDir, t.RunPath)
	if pathExists(toolchainPath) {
		return nil
	}

	// Download to fixed dir.
	downloaded, err := io.Download(t.Url, DownloadDir)
	if err != nil {
		return fmt.Errorf("%s: download toolchain failed: %w", t.Url, err)
	}

	// Extract archive file.
	fileName := filepath.Base(t.Url)
	folderName := strings.TrimSuffix(fileName, ".tar.gz")
	extractPath := filepath.Join(DownloadDir, folderName)
	if err := io.Extract(downloaded, extractPath); err != nil {
		return fmt.Errorf("%s: extract toolchain failed: %w", downloaded, err)
	}

	fmt.Printf("[✔] -------- %s.\n\n", filepath.Base(t.Url))
	return nil
}
