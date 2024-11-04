package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	PlatformsDir = "conf/platforms"
	ToolsDir     = "conf/tools"

	WorkspaceDir  = "workspace"
	DownloadDir   = "workspace/downloads"
	ToolchainFile = "buildenv.cmake"
)

type BuildEnv struct {
	RootFS    RootFS    `json:"rootfs"`
	Toolchain Toolchain `json:"toolchain"`
	Tools     []string  `json:"tools"`
	Packages  []string  `json:"packages"`
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

func (b BuildEnv) Verify(resRepoUrl string, onlyFields bool) error {
	if resRepoUrl == "" {
		return fmt.Errorf("repoUrl is empty")
	}

	if err := b.RootFS.Verify(resRepoUrl, onlyFields); err != nil {
		return fmt.Errorf("buildenv.rootfs error: %w", err)
	}

	if err := b.Toolchain.Verify(resRepoUrl, onlyFields); err != nil {
		return fmt.Errorf("buildenv.toolchain error: %w", err)
	}

	for _, item := range b.Tools {
		var tool Tool
		if err := tool.Verify(resRepoUrl, item, onlyFields); err != nil {
			return fmt.Errorf("buildenv.tools[%s] error: %w", item, err)
		}
	}

	return nil
}

func (b BuildEnv) CreateToolchainFile(outputDir string) (string, error) {
	var toolchain, environment strings.Builder

	// Verify buildenv during configuration.
	toolchain.WriteString("\n# Verify buildenv during configuration.\n")
	toolchain.WriteString("set(HOME_DIR \"${CMAKE_CURRENT_LIST_DIR}/..\")\n")
	toolchain.WriteString("set(BUILDENV_EXECUTABLE \"${HOME_DIR}/buildenv\")\n")
	toolchain.WriteString("execute_process(\n")
	toolchain.WriteString("\tCOMMAND ${BUILDENV_EXECUTABLE} --verify\n")
	toolchain.WriteString("\tWORKING_DIRECTORY ${HOME_DIR}\n")
	toolchain.WriteString(")\n")

	// Set toolchain platform infos.
	toolchain.WriteString("\n# Set toolchain platform infos.\n")
	toolchain.WriteString(fmt.Sprintf("set(CMAKE_SYSTEM_NAME \"%s\")\n", b.Toolchain.ToolChainVars.CMAKE_SYSTEM_NAME))
	toolchain.WriteString(fmt.Sprintf("set(CMAKE_SYSTEM_PROCESSOR \"%s\")\n", b.Toolchain.ToolChainVars.CMAKE_SYSTEM_PROCESSOR))

	// Set sysroot for cross-compile.
	if err := b.writeRootFS(&toolchain, &environment); err != nil {
		return "", err
	}

	// Set toolchain for cross-compile.
	if err := b.writeToolchain(&toolchain, &environment); err != nil {
		return "", err
	}

	// Search programs in the host environment.
	toolchain.WriteString("\n# Search programs in the host environment.\n")
	toolchain.WriteString("set(CMAKE_FIND_ROOT_PATH_MODE_PROGRAM NEVER)\n")

	// Search libraries and headers in the target environment.
	toolchain.WriteString("\n# Search libraries and headers in the target environment.\n")
	toolchain.WriteString("set(CMAKE_FIND_ROOT_PATH_MODE_LIBRARY ONLY)\n")
	toolchain.WriteString("set(CMAKE_FIND_ROOT_PATH_MODE_INCLUDE ONLY)\n")
	toolchain.WriteString("set(CMAKE_FIND_ROOT_PATH_MODE_PACKAGE ONLY)\n")

	// Create the output directory if it doesn't exist.
	if err := os.MkdirAll(outputDir, os.ModeDir|os.ModePerm); err != nil {
		return "", err
	}

	// Write the modified content to the output file.
	toolchainPath := filepath.Join(outputDir, ToolchainFile)
	if err := os.WriteFile(toolchainPath, []byte(toolchain.String()), os.ModePerm); err != nil {
		return "", err
	}

	// Write environment variables to the file.
	environmentPath := filepath.Join(outputDir, "buildenv.sh")
	if err := os.WriteFile(environmentPath, []byte(environment.String()), os.ModePerm); err != nil {
		return "", err
	}

	// Set permissions for the file: rwxr-xr-x
	if err := os.Chmod(environmentPath, 0755); err != nil {
		log.Fatalf("Error setting permissions: %v", err)
	}

	return toolchainPath, nil
}

func (b *BuildEnv) writeRootFS(toolchain, environment *strings.Builder) error {
	if !b.RootFS.None {
		rootFSPath := b.RootFS.AbsolutePath()

		toolchain.WriteString("\n# Set sysroot for cross-compile.\n")
		toolchain.WriteString(fmt.Sprintf("set(CMAKE_SYSROOT \"%s\")\n", rootFSPath))
		toolchain.WriteString("list(APPEND CMAKE_FIND_ROOT_PATH \"${CMAKE_SYSROOT}\")\n")
		toolchain.WriteString("list(APPEND CMAKE_PREFIX_PATH NEVER \"${CMAKE_SYSROOT}\")\n")

		toolchain.WriteString("\n# Set pkg-config path for cross-compile.\n")
		toolchain.WriteString("set(ENV{PKG_CONFIG_SYSROOT_DIR} \"${CMAKE_SYSROOT}\")\n")

		// Replace the path with the workspace directory.
		for i, path := range b.RootFS.EnvVars.PKG_CONFIG_PATH {
			fullPath := filepath.Join(WorkspaceDir, path)
			absPath, err := filepath.Abs(fullPath)
			if err != nil {
				return fmt.Errorf("cannot get absolute path: %s", fullPath)
			}

			b.RootFS.EnvVars.PKG_CONFIG_PATH[i] = absPath
		}
		toolchain.WriteString(fmt.Sprintf("set(ENV{PKG_CONFIG_PATH} \"%s\")\n", strings.Join(b.RootFS.EnvVars.PKG_CONFIG_PATH, ":")))

		// Set environment variables for makefile project.
		environment.WriteString(fmt.Sprintf("export SYSROOT=%s\n", rootFSPath))
		environment.WriteString("export PKG_CONFIG_SYSROOT_DIR=${SYSROOT}\n")
		environment.WriteString(fmt.Sprintf("export PKG_CONFIG_PATH=%s\n\n", strings.Join(b.RootFS.EnvVars.PKG_CONFIG_PATH, ":")))
	}

	return nil
}

func (b *BuildEnv) writeToolchain(toolchain, environment *strings.Builder) error {
	toolchain.WriteString("\n# Set toolchain for cross-compile.\n")
	toolchainBinPath := filepath.Join(WorkspaceDir, b.Toolchain.Path)
	absToolchainBinPath, err := filepath.Abs(toolchainBinPath)
	if err != nil {
		return fmt.Errorf("cannot get absolute path of toolchain path: %s", toolchainBinPath)
	}

	toolchain.WriteString(fmt.Sprintf("set(_TOOLCHAIN_BIN_PATH		\"%s\")\n", absToolchainBinPath))

	writeIfNotEmpty := func(content, env, value string) {
		if value != "" {
			// Set toolchain variables.
			toolchain.WriteString(fmt.Sprintf("set(%s%s\")\n", content, value))

			// Set environment variables for makefile project.
			environment.WriteString(fmt.Sprintf("export %s=${TOOLCHAIN_BIN_PATH}/%s\n", env, value))
		}
	}

	environment.WriteString(fmt.Sprintf("export TOOLCHAIN_BIN_PATH=%s\n", absToolchainBinPath))

	writeIfNotEmpty("CMAKE_C_COMPILER 		\"${_TOOLCHAIN_BIN_PATH}/", "CC", b.Toolchain.EnvVars.CC)
	writeIfNotEmpty("CMAKE_CXX_COMPILER		\"${_TOOLCHAIN_BIN_PATH}/", "CXX", b.Toolchain.EnvVars.CXX)
	writeIfNotEmpty("CMAKE_Fortran_COMPILER	\"${_TOOLCHAIN_BIN_PATH}/", "FC", b.Toolchain.EnvVars.FC)
	writeIfNotEmpty("CMAKE_RANLIB 			\"${_TOOLCHAIN_BIN_PATH}/", "RANLIB", b.Toolchain.EnvVars.RANLIB)
	writeIfNotEmpty("CMAKE_AR 				\"${_TOOLCHAIN_BIN_PATH}/", "AR", b.Toolchain.EnvVars.AR)
	writeIfNotEmpty("CMAKE_LINKER 			\"${_TOOLCHAIN_BIN_PATH}/", "LD", b.Toolchain.EnvVars.LD)
	writeIfNotEmpty("CMAKE_NM 				\"${_TOOLCHAIN_BIN_PATH}/", "NM", b.Toolchain.EnvVars.NM)
	writeIfNotEmpty("CMAKE_OBJDUMP 			\"${_TOOLCHAIN_BIN_PATH}/", "OBJDUMP", b.Toolchain.EnvVars.OBJDUMP)
	writeIfNotEmpty("CMAKE_STRIP 			\"${_TOOLCHAIN_BIN_PATH}/", "STRIP", b.Toolchain.EnvVars.STRIP)

	return nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	return !os.IsNotExist(err)
}
