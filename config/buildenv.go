package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	PlatformDir   = "conf/platform"
	WorkspaceDir  = "workspace"
	DownloadDir   = "workspace/downloads"
	ToolchainDir  = "workspace/toolchain"
	RootFSDir     = "workspace/rootfs"
	ToolchainFile = "buildenv.cmake"
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

func (b BuildEnv) Verify(onlyFields bool) error {
	if b.Host == "" {
		return fmt.Errorf("buildenv.hostUrl is empty")
	}

	if err := b.RootFS.Verify(b.Host, onlyFields); err != nil {
		return fmt.Errorf("buildenv.rootfs error: %w", err)
	}

	if err := b.Toolchain.Verify(b.Host, onlyFields); err != nil {
		return fmt.Errorf("buildenv.toolchain error: %w", err)
	}

	return nil
}

func (b BuildEnv) CreateToolchainFile(outputDir string) (string, error) {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("set(CMAKE_SYSTEM_NAME \"%s\")\n", b.Toolchain.ToolChainVars.CMAKE_SYSTEM_NAME))
	builder.WriteString(fmt.Sprintf("set(CMAKE_SYSTEM_PROCESSOR \"%s\")\n", b.Toolchain.ToolChainVars.CMAKE_SYSTEM_PROCESSOR))

	// Set sysroot for cross-compile.
	if !b.RootFS.None {
		builder.WriteString("\n# Set sysroot for cross-compile.\n")
		builder.WriteString(fmt.Sprintf("set(CMAKE_SYSROOT \"%s\")\n", b.RootFS.AbsolutePath()))
		builder.WriteString("list(APPEND CMAKE_FIND_ROOT_PATH \"${CMAKE_SYSROOT}\")\n")
		builder.WriteString("list(APPEND CMAKE_PREFIX_PATH NEVER \"${CMAKE_SYSROOT}\")\n")

		builder.WriteString("\n# Set pkg-config path for cross-compile.\n")
		builder.WriteString("set(ENV{PKG_CONFIG_SYSROOT_DIR} \"${CMAKE_SYSROOT}\")\n")

		// Replace the path with the workspace directory.
		for i, path := range b.RootFS.EnvVars.PKG_CONFIG_PATH {
			fullPath := filepath.Join(RootFSDir, path)
			absPath, err := filepath.Abs(fullPath)
			if err != nil {
				return "", fmt.Errorf("cannot get absolute path: %s", fullPath)
			}

			b.RootFS.EnvVars.PKG_CONFIG_PATH[i] = absPath
		}
		builder.WriteString(fmt.Sprintf("set(ENV{PKG_CONFIG_PATH} \"%s\")\n", strings.Join(b.RootFS.EnvVars.PKG_CONFIG_PATH, ";")))
	}

	// Set toolchain for cross-compile.
	builder.WriteString("\n# Set toolchain for cross-compile.\n")
	toolchainBinPath := filepath.Join(WorkspaceDir, b.Toolchain.Path)
	absToolchainBinPath, err := filepath.Abs(toolchainBinPath)
	if err != nil {
		return "", fmt.Errorf("cannot get absolute path of toolchain path: %s", toolchainBinPath)
	}
	builder.WriteString(fmt.Sprintf("set(_TOOLCHAIN_BIN_PATH 	\"%s\")\n", absToolchainBinPath))

	builder.WriteString(fmt.Sprintf("set(CMAKE_C_COMPILER 		\"${_TOOLCHAIN_BIN_PATH}/%s\")\n", b.Toolchain.EnvVars.CC))
	builder.WriteString(fmt.Sprintf("set(CMAKE_CXX_COMPILER		\"${_TOOLCHAIN_BIN_PATH}/%s\")\n", b.Toolchain.EnvVars.CXX))
	builder.WriteString(fmt.Sprintf("set(CMAKE_Fortran_COMPILER	\"${_TOOLCHAIN_BIN_PATH}/%s\")\n", b.Toolchain.EnvVars.FC))
	builder.WriteString(fmt.Sprintf("set(CMAKE_RANLIB 			\"${_TOOLCHAIN_BIN_PATH}/%s\")\n", b.Toolchain.EnvVars.RANLIB))
	builder.WriteString(fmt.Sprintf("set(CMAKE_AR 				\"${_TOOLCHAIN_BIN_PATH}/%s\")\n", b.Toolchain.EnvVars.AR))
	builder.WriteString(fmt.Sprintf("set(CMAKE_LINKER 			\"${_TOOLCHAIN_BIN_PATH}/%s\")\n", b.Toolchain.EnvVars.LD))
	builder.WriteString(fmt.Sprintf("set(CMAKE_NM 				\"${_TOOLCHAIN_BIN_PATH}/%s\")\n", b.Toolchain.EnvVars.NM))
	builder.WriteString(fmt.Sprintf("set(CMAKE_OBJDUMP 			\"${_TOOLCHAIN_BIN_PATH}/%s\")\n", b.Toolchain.EnvVars.OBJDUMP))
	builder.WriteString(fmt.Sprintf("set(CMAKE_STRIP 			\"${_TOOLCHAIN_BIN_PATH}/%s\")\n", b.Toolchain.EnvVars.STRIP))

	// Search programs in the host environment.
	builder.WriteString("\n# Search programs in the host environment.\n")
	builder.WriteString("set(CMAKE_FIND_ROOT_PATH_MODE_PROGRAM NEVER)\n")

	// Search libraries and headers in the target environment.
	builder.WriteString("\n# Search libraries and headers in the target environment.\n")
	builder.WriteString("set(CMAKE_FIND_ROOT_PATH_MODE_LIBRARY ONLY)\n")
	builder.WriteString("set(CMAKE_FIND_ROOT_PATH_MODE_INCLUDE ONLY)\n")
	builder.WriteString("set(CMAKE_FIND_ROOT_PATH_MODE_PACKAGE ONLY)\n")

	// Create the output directory if it doesn't exist.
	if err := os.MkdirAll(outputDir, os.ModeDir|os.ModePerm); err != nil {
		return "", err
	}

	// Write the modified content to the output file.
	filePath := filepath.Join(outputDir, ToolchainFile)
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
