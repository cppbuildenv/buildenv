package config

import (
	"buildenv/pkg/io"
	"fmt"
	"path/filepath"
)

type Toolchain struct {
	Url           string          `json:"url"`
	ExtractPath   string          `json:"extract_path"`
	RuntimePath   string          `json:"runtime_path"`
	EnvVars       ToolchainEnvVar `json:"env_vars"`
	ToolChainVars ToolChainVars   `json:"toolchain_vars"`
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

	if t.ExtractPath == "" {
		return fmt.Errorf("toolchain.extract_path is empty")
	}

	if t.RuntimePath == "" {
		return fmt.Errorf("toolchain.runtime_path is empty")
	}

	if t.EnvVars.CC == "" {
		return fmt.Errorf("toolchain.env.CC is empty")
	}

	if t.EnvVars.CXX == "" {
		return fmt.Errorf("toolchain.env.CXX is empty")
	}

	if t.ToolChainVars.CMAKE_SYSTEM_NAME == "" {
		return fmt.Errorf("toolchain.toolchain_vars.CMAKE_SYSTEM_NAME is empty")
	}

	if t.ToolChainVars.CMAKE_SYSTEM_PROCESSOR == "" {
		return fmt.Errorf("toolchain.toolchain_vars.CMAKE_SYSTEM_PROCESSOR is empty")
	}

	if !checkAndRepiar {
		return nil
	}

	return t.checkAndRepair()
}

func (t Toolchain) checkAndRepair() error {
	toolchainPath := filepath.Join(WorkspaceDir, t.RuntimePath)
	if pathExists(toolchainPath) {
		return nil
	}

	// Download to fixed dir.
	downloaded, err := io.Download(t.Url, DownloadDir)
	if err != nil {
		return fmt.Errorf("%s: download toolchain failed: %w", t.Url, err)
	}

	// Extract to `extract_path`
	extractDir := filepath.Join(WorkspaceDir, t.ExtractPath)
	if err := io.Extract(downloaded, extractDir); err != nil {
		return fmt.Errorf("%s: extract toolchain failed: %w", downloaded, err)
	}

	fmt.Printf("[✔] ---- toolchain of platform setup success.\n\n")
	return nil
}
