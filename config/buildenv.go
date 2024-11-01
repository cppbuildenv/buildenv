package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const PlatformDir = "conf/platform"

// ==============================  buildenv ============================== //
type BuildEnv struct {
	Host      string    `json:"host"`
	RootFS    RootFS    `json:"rootfs"`
	Toolchain Toolchain `json:"toolchain"`
}

func (b BuildEnv) Verify() error {
	if b.Host == "" {
		return fmt.Errorf("buildenv.hostUrl is empty")
	}

	if err := b.RootFS.Verify(); err != nil {
		return fmt.Errorf("buildenv.rootfs error: %w", err)
	}

	if err := b.Toolchain.Verify(); err != nil {
		return fmt.Errorf("buildenv.toolchain error: %w", err)
	}

	return nil
}

func (b *BuildEnv) Read(filePath string) error {
	// Check if platform file exists
	if !pathExists(filePath) {
		return fmt.Errorf("platform file not exists: %s", filePath)
	}

	// Read conf/buildenv.json
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bytes, b); err != nil {
		return fmt.Errorf("%s read error: %w", filePath, err)
	}

	return nil
}

func (b BuildEnv) Write(filePath string) error {
	// Create empty array for empty field.
	if len(b.RootFS.EnvVars.PKG_CONFIG_PATH) == 0 {
		b.RootFS.EnvVars.PKG_CONFIG_PATH = []string{}
	}

	bytes, err := json.MarshalIndent(b, "", "    ")
	if err != nil {
		return err
	}

	// Check if conf/buildenv.json exists
	if pathExists(filePath) {
		return fmt.Errorf("[%s] is already exists", filePath)
	}

	// Makesure the parent directory exists.
	parentDir := filepath.Dir(filePath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filePath, bytes, os.ModePerm)
}

// ==============================  rootfs ============================== //
type RootFS struct {
	Url     string    `json:"url"`
	EnvVars RootFSEnv `json:"env_vars"`
}

type RootFSEnv struct {
	SYSROOT                string   `json:"SYSROOT"`
	PKG_CONFIG_SYSROOT_DIR string   `json:"PKG_CONFIG_SYSROOT_DIR"`
	PKG_CONFIG_PATH        []string `json:"PKG_CONFIG_PATH"`
}

func (r RootFS) Verify() error {
	if r.Url == "" {
		return fmt.Errorf("rootfs.url is empty")
	}

	if r.EnvVars.SYSROOT == "" {
		return fmt.Errorf("rootfs.env.SYSROOT is empty")
	}

	if r.EnvVars.PKG_CONFIG_SYSROOT_DIR == "" {
		return fmt.Errorf("rootfs.env.PKG_CONFIG_SYSROOT_DIR is empty")
	}

	if len(r.EnvVars.PKG_CONFIG_PATH) == 0 {
		return fmt.Errorf("rootfs.env.PKG_CONFIG_PATH is empty")
	}

	return nil
}

// ==============================  toolchain ============================== //
type Toolchain struct {
	Url           string          `json:"url"`
	Path          string          `json:"path"`
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

func (t Toolchain) Verify() error {
	if t.Url == "" {
		return fmt.Errorf("toolchain.url is empty")
	}

	if t.Path == "" {
		return fmt.Errorf("toolchain.path is empty")
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

	return nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	return !os.IsNotExist(err)
}
