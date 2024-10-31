package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ==============================  buildenv ============================== //
type BuildEnv struct {
	HostUrl   string    `json:"hostUrl"`
	RootFS    RootFS    `json:"rootfs"`
	Toolchain Toolchain `json:"toolchain"`
}

func (b BuildEnv) Verify() error {
	if b.HostUrl == "" {
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

func (b BuildEnv) Write(platformName string, forcely bool) error {
	// Create empty array for empty field.
	if len(b.RootFS.Env.PKG_CONFIG_PATH) == 0 {
		b.RootFS.Env.PKG_CONFIG_PATH = []string{}
	}

	bytes, err := json.MarshalIndent(b, "", "    ")
	if err != nil {
		return err
	}

	// Check if conf/buildenv.json exists
	filePath := fmt.Sprintf("conf/platform/%s.json", platformName)
	if pathExists(filePath) {
		if !forcely {
			return fmt.Errorf("it's already exists, but you can create with -f to overwrite")
		}
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
	Url string    `json:"url"`
	Env RootFSEnv `json:"env"`
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

	if r.Env.SYSROOT == "" {
		return fmt.Errorf("rootfs.env.SYSROOT is empty")
	}

	if r.Env.PKG_CONFIG_SYSROOT_DIR == "" {
		return fmt.Errorf("rootfs.env.PKG_CONFIG_SYSROOT_DIR is empty")
	}

	if len(r.Env.PKG_CONFIG_PATH) == 0 {
		return fmt.Errorf("rootfs.env.PKG_CONFIG_PATH is empty")
	}

	return nil
}

// ==============================  toolchain ============================== //
type Toolchain struct {
	Url  string       `json:"url"`
	Path string       `json:"path"`
	Env  ToolchainEnv `json:"env"`
}

type ToolchainEnv struct {
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

func (t Toolchain) Verify() error {
	if t.Url == "" {
		return fmt.Errorf("toolchain.url is empty")
	}

	if t.Path == "" {
		return fmt.Errorf("toolchain.path is empty")
	}

	if t.Env.CC == "" {
		return fmt.Errorf("toolchain.env.CC is empty")
	}

	if t.Env.CXX == "" {
		return fmt.Errorf("toolchain.env.CXX is empty")
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
