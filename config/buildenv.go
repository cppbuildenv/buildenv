package config

import (
	"buildenv/pkg/io"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	PlatformDir = "conf/platform"
	DownloadDir = "downloads"
)

type BuildEnv struct {
	Host      string    `json:"host"`
	RootFS    RootFS    `json:"rootfs"`
	Toolchain Toolchain `json:"toolchain"`
}

func (b *BuildEnv) Read(filePath string) error {
	// Check if platform file exists.
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

	// Check if conf/buildenv.json exists.
	if pathExists(filePath) {
		return fmt.Errorf("[%s] is already exists", filePath)
	}

	// Make sure the parent directory exists.
	parentDir := filepath.Dir(filePath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filePath, bytes, os.ModePerm)
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

func (b BuildEnv) CheckIntegrity() error {
	rootfsPath := filepath.Join(DownloadDir, b.RootFS.Path)
	if !pathExists(rootfsPath) {
		fullUrl, err := url.JoinPath(b.Host, b.RootFS.Url)
		if err != nil {
			return fmt.Errorf("buildenv.rootfs.url error: %w", err)
		}

		io.Download(fullUrl, DownloadDir)
	}
	return nil
}

func (b BuildEnv) CreateToolchainFile(outputDir string) (string, error) {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("set(CMAKE_SYSTEM_NAME \"%s\")\n", b.Toolchain.ToolChainVars.CMAKE_SYSTEM_NAME))
	builder.WriteString(fmt.Sprintf("set(CMAKE_SYSTEM_PROCESSOR \"%s\")\n", b.Toolchain.ToolChainVars.CMAKE_SYSTEM_PROCESSOR))

	if !b.RootFS.None {
		builder.WriteString("\n# Set sysroot for cross-compile.\n")
		builder.WriteString(fmt.Sprintf("set(CMAKE_SYSROOT \"%s\")\n", b.RootFS.Path))
		builder.WriteString(fmt.Sprintf("list(APPEND CMAKE_FIND_ROOT_PATH \"%s\")\n", b.RootFS.Path))
		builder.WriteString(fmt.Sprintf("list(APPEND CMAKE_PREFIX_PATH NEVER \"%s\")\n", b.RootFS.Path))

		builder.WriteString("\n# Set pkg-config path for cross-compile.\n")
		builder.WriteString(fmt.Sprintf("set(ENV{PKG_CONFIG_SYSROOT_DIR} \"%s\")\n", b.RootFS.EnvVars.PKG_CONFIG_SYSROOT_DIR))
		builder.WriteString(fmt.Sprintf("set(ENV{PKG_CONFIG_PATH} \"%s\")\n", strings.Join(b.RootFS.EnvVars.PKG_CONFIG_PATH, ";")))
	}

	builder.WriteString("\n# Set toolchain for cross-compile.\n")
	builder.WriteString(fmt.Sprintf("set(CMAKE_C_COMPILER \"%s\")\n", b.Toolchain.EnvVars.CC))
	builder.WriteString(fmt.Sprintf("set(CMAKE_CXX_COMPILER \"%s\")\n", b.Toolchain.EnvVars.CXX))
	builder.WriteString(fmt.Sprintf("set(CMAKE_Fortran_COMPILER \"%s\")\n", b.Toolchain.EnvVars.FC))
	builder.WriteString(fmt.Sprintf("set(CMAKE_RANLIB \"%s\")\n", b.Toolchain.EnvVars.RANLIB))
	builder.WriteString(fmt.Sprintf("set(CMAKE_AR \"%s\")\n", b.Toolchain.EnvVars.AR))
	builder.WriteString(fmt.Sprintf("set(CMAKE_LINKER \"%s\")\n", b.Toolchain.EnvVars.LD))
	builder.WriteString(fmt.Sprintf("set(CMAKE_NM \"%s\")\n", b.Toolchain.EnvVars.NM))
	builder.WriteString(fmt.Sprintf("set(CMAKE_OBJDUMP \"%s\")\n", b.Toolchain.EnvVars.OBJDUMP))
	builder.WriteString(fmt.Sprintf("set(CMAKE_STRIP \"%s\")\n", b.Toolchain.EnvVars.STRIP))

	builder.WriteString("\n# Search programs in the host environment.\n")
	builder.WriteString("set(CMAKE_FIND_ROOT_PATH_MODE_PROGRAM NEVER)\n")

	builder.WriteString("\n# Search libraries and headers in the target environment.\n")
	builder.WriteString("set(CMAKE_FIND_ROOT_PATH_MODE_LIBRARY ONLY)\n")
	builder.WriteString("set(CMAKE_FIND_ROOT_PATH_MODE_INCLUDE ONLY)\n")
	builder.WriteString("set(CMAKE_FIND_ROOT_PATH_MODE_PACKAGE ONLY)\n")

	// Write the modified content to the output file.
	filePath := filepath.Join(outputDir, "toolchain_buildenv.cmake")
	if err := os.WriteFile(filePath, []byte(builder.String()), os.ModePerm); err != nil {
		return "", err
	}

	return filePath, nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	return !os.IsNotExist(err)
}
