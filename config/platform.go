package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type PlatformCallbacks interface {
	OnCreatePlatform(platformName string) error
	OnSelectPlatform(platformName string) error
}

type Platform struct {
	RootFS       *RootFS    `json:"rootfs"`
	Toolchain    *Toolchain `json:"toolchain"`
	Tools        []string   `json:"tools"`
	Dependencies []string   `json:"dependencies"`

	// Internal fields.
	platformName string `json:"-"`
}

func (b *Platform) Read(platformPath string) error {
	// Check if platform file exists.
	if !pathExists(platformPath) {
		return fmt.Errorf("platform file not exists: %s", platformPath)
	}

	// Read conf/buildenv.json
	bytes, err := os.ReadFile(platformPath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bytes, b); err != nil {
		return fmt.Errorf("read error: %w", err)
	}

	// Set values of internal fields.
	b.platformName = strings.TrimSuffix(filepath.Base(platformPath), ".json")
	return nil
}

func (b Platform) Write(filePath string) error {
	// Create empty array for empty field.
	b.RootFS = new(RootFS)
	b.Toolchain = new(Toolchain)

	if len(b.RootFS.EnvVars.PKG_CONFIG_PATH) == 0 {
		b.RootFS.EnvVars.PKG_CONFIG_PATH = []string{}
	}
	if len(b.Tools) == 0 {
		b.Tools = []string{}
	}
	if len(b.Dependencies) == 0 {
		b.Dependencies = []string{}
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

func (b Platform) Verify(args VerifyArgs) error {
	// RootFS maybe nil when platform is native.
	if b.RootFS != nil {
		if err := b.RootFS.Verify(args); err != nil {
			return fmt.Errorf("buildenv.rootfs error: %w", err)
		}
	}

	// Toolchain maybe nil when platform is native.
	if b.Toolchain != nil {
		if err := b.Toolchain.Verify(args); err != nil {
			return fmt.Errorf("buildenv.toolchain error: %w", err)
		}
	}

	// Verify tools.
	for _, item := range b.Tools {
		toolpath := filepath.Join(Dirs.ToolDir, item+".json")
		var tool Tool

		if err := tool.Init(toolpath); err != nil {
			return fmt.Errorf("buildenv.tools[%s] read error: %w", item, err)
		}

		if err := tool.Verify(args); err != nil {
			return fmt.Errorf("buildenv.tools[%s] verify error: %w", item, err)
		}
	}

	// Verify dependencies.
	for _, item := range b.Dependencies {
		portPath := filepath.Join(Dirs.PortDir, item+".json")
		var port Port
		if err := port.Init(portPath, b.platformName, args.BuildType); err != nil {
			return fmt.Errorf("buildenv.dependencies[%s] read error: %w", item, err)
		}

		if err := port.Verify(args); err != nil {
			return fmt.Errorf("buildenv.dependencies[%s] verify error: %w", item, err)
		}
	}

	return nil
}

func (b Platform) CreateToolchainFile(scriptDir string) (string, error) {
	var toolchain, environment strings.Builder

	// Set default CMAKE_BUILD_TYPE.
	toolchain.WriteString("# Set default CMAKE_BUILD_TYPE.\n")
	toolchain.WriteString("if(NOT CMAKE_BUILD_TYPE)\n")
	toolchain.WriteString("\tset(CMAKE_BUILD_TYPE \"Release\")\n")
	toolchain.WriteString("endif()\n")

	// Verify buildenv during configuration.
	toolchain.WriteString("\n# Verify buildenv during configuration.\n")
	toolchain.WriteString("set(HOME_DIR \"${CMAKE_CURRENT_LIST_DIR}/..\")\n")
	toolchain.WriteString("set(BUILDENV_EXECUTABLE \"${HOME_DIR}/buildenv\")\n")
	toolchain.WriteString("execute_process(\n")
	toolchain.WriteString("\tCOMMAND ${BUILDENV_EXECUTABLE} -verify -silent -build_type=${CMAKE_BUILD_TYPE}\n")
	toolchain.WriteString("\tWORKING_DIRECTORY ${HOME_DIR}\n")
	toolchain.WriteString(")\n")

	// Set toolchain platform infos.
	if b.Toolchain != nil {
		toolchain.WriteString("\n# Set toolchain platform infos.\n")
		toolchain.WriteString(fmt.Sprintf("set(CMAKE_SYSTEM_NAME \"%s\")\n", b.Toolchain.SystemName))
		toolchain.WriteString(fmt.Sprintf("set(CMAKE_SYSTEM_PROCESSOR \"%s\")\n", b.Toolchain.SystemProcessor))
	}

	// Set sysroot for cross-compile.
	if err := b.writeRootFS(&toolchain, &environment); err != nil {
		return "", err
	}

	// Set toolchain for cross-compile.
	if err := b.writeToolchain(&toolchain, &environment); err != nil {
		return "", err
	}

	// Set tools for cross-compile.
	if err := b.writeTools(&toolchain, &environment); err != nil {
		return "", err
	}

	// Create the output directory if it doesn't exist.
	if err := os.MkdirAll(scriptDir, os.ModeDir|os.ModePerm); err != nil {
		return "", err
	}

	// Write toolchain file.
	toolchainPath := filepath.Join(scriptDir, "buildenv.cmake")
	if err := os.WriteFile(toolchainPath, []byte(toolchain.String()), os.ModePerm); err != nil {
		return "", err
	}

	// Write environment file.
	environmentPath := filepath.Join(scriptDir, "buildenv.sh")
	if err := os.WriteFile(environmentPath, []byte(environment.String()), os.ModePerm); err != nil {
		return "", err
	}

	// Grant executable permission to the file: rwxr-xr-x
	if err := os.Chmod(environmentPath, 0755); err != nil {
		log.Fatalf("Error setting permissions: %v", err)
	}

	return toolchainPath, nil
}

func (b *Platform) writeRootFS(toolchain, environment *strings.Builder) error {
	// Skip if rootfs is nil.
	if b.RootFS == nil {
		return nil
	}

	// Generate rootfs description.
	rootfsPath := filepath.Join(Dirs.DownloadDir, b.RootFS.RunPath)
	absRootFSPath, err := filepath.Abs(rootfsPath)
	if err != nil {
		panic(fmt.Sprintf("cannot get absolute path: %s", rootfsPath))
	}

	toolchain.WriteString("\n# Set sysroot for cross-compile.\n")
	toolchain.WriteString(fmt.Sprintf("set(CMAKE_SYSROOT \"%s\")\n", absRootFSPath))
	toolchain.WriteString("list(APPEND CMAKE_FIND_ROOT_PATH \"${CMAKE_SYSROOT}\")\n")
	toolchain.WriteString("list(APPEND CMAKE_PREFIX_PATH \"${CMAKE_SYSROOT}\")\n")

	// Search programs in the host environment.
	toolchain.WriteString("\n# Search programs in the host environment.\n")
	toolchain.WriteString("set(CMAKE_FIND_ROOT_PATH_MODE_PROGRAM NEVER)\n")

	// Search libraries and headers in the target environment.
	toolchain.WriteString("\n# Search libraries and headers in the target environment.\n")
	toolchain.WriteString("set(CMAKE_FIND_ROOT_PATH_MODE_LIBRARY ONLY)\n")
	toolchain.WriteString("set(CMAKE_FIND_ROOT_PATH_MODE_INCLUDE ONLY)\n")
	toolchain.WriteString("set(CMAKE_FIND_ROOT_PATH_MODE_PACKAGE ONLY)\n")

	toolchain.WriteString("\n# Set pkg-config path for cross-compile.\n")
	toolchain.WriteString("set(ENV{PKG_CONFIG_SYSROOT_DIR} \"${CMAKE_SYSROOT}\")\n")

	// Replace the path with the workspace directory.
	for i, path := range b.RootFS.EnvVars.PKG_CONFIG_PATH {
		fullPath := filepath.Join(Dirs.DownloadDir, path)
		absPath, err := filepath.Abs(fullPath)
		if err != nil {
			return fmt.Errorf("cannot get absolute path: %s", fullPath)
		}

		b.RootFS.EnvVars.PKG_CONFIG_PATH[i] = absPath
	}
	toolchain.WriteString(fmt.Sprintf("set(ENV{PKG_CONFIG_PATH} \"%s\")\n", strings.Join(b.RootFS.EnvVars.PKG_CONFIG_PATH, ":")))

	// Set environment variables for makefile project.
	environment.WriteString("\n# Set rootfs for cross compile.\n")
	environment.WriteString(fmt.Sprintf("export SYSROOT=%s\n", absRootFSPath))
	environment.WriteString("export PATH=${SYSROOT}:${PATH}\n")
	environment.WriteString("export PKG_CONFIG_SYSROOT_DIR=${SYSROOT}\n")
	environment.WriteString(fmt.Sprintf("export PKG_CONFIG_PATH=%s\n\n", strings.Join(b.RootFS.EnvVars.PKG_CONFIG_PATH, ":")))

	// Make sure the toolchain is in the PATH of current process.
	os.Setenv("SYSROOT", absRootFSPath)
	os.Setenv("PKG_CONFIG_SYSROOT_DIR", absRootFSPath)
	os.Setenv("PKG_CONFIG_PATH", strings.Join(b.RootFS.EnvVars.PKG_CONFIG_PATH, ":"))
	os.Setenv("PATH", fmt.Sprintf("%s%c%s", absRootFSPath, os.PathListSeparator, os.Getenv("PATH")))

	return nil
}

func (b *Platform) writeToolchain(toolchain, environment *strings.Builder) error {
	// Skip if toolchain is nil.
	if b.Toolchain == nil {
		return nil
	}

	// Generate toolchain description.
	toolchain.WriteString("\n# Set toolchain for cross-compile.\n")
	toolchainPath := filepath.Join(Dirs.DownloadDir, b.Toolchain.RunPath)
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

			// Make sure the tool is in the PATH of current process.
			os.Setenv(strings.TrimSpace(env), value)
		}
	}

	environment.WriteString("# Set toolchain for cross compile.\n")
	environment.WriteString(fmt.Sprintf("export TOOLCHAIN_PATH=%s\n", absToolchainPath))
	environment.WriteString("export PATH=${TOOLCHAIN_PATH}:${PATH}\n\n")

	// Make sure the toolchain is in the PATH of current process.
	os.Setenv("PATH", fmt.Sprintf("%s%c%s", absToolchainPath, os.PathListSeparator, os.Getenv("PATH")))

	writeIfNotEmpty("CMAKE_C_COMPILER 		", "CC", b.Toolchain.EnvVars.CC)
	writeIfNotEmpty("CMAKE_CXX_COMPILER		", "CXX", b.Toolchain.EnvVars.CXX)
	writeIfNotEmpty("CMAKE_Fortran_COMPILER	", "FC", b.Toolchain.EnvVars.FC)
	writeIfNotEmpty("CMAKE_RANLIB 			", "RANLIB", b.Toolchain.EnvVars.RANLIB)
	writeIfNotEmpty("CMAKE_AR 				", "AR", b.Toolchain.EnvVars.AR)
	writeIfNotEmpty("CMAKE_LINKER 			", "LD", b.Toolchain.EnvVars.LD)
	writeIfNotEmpty("CMAKE_NM 				", "NM", b.Toolchain.EnvVars.NM)
	writeIfNotEmpty("CMAKE_OBJDUMP 			", "OBJDUMP", b.Toolchain.EnvVars.OBJDUMP)
	writeIfNotEmpty("CMAKE_STRIP 			", "STRIP", b.Toolchain.EnvVars.STRIP)

	return nil
}

func (b *Platform) writeTools(toolchain, environment *strings.Builder) error {
	toolchain.WriteString("\n# Append `run_path` of tools into $PATH.\n")
	environment.WriteString("\n# Append `run_path` of tools into $PATH.\n")

	for _, item := range b.Tools {
		toolPath := filepath.Join(Dirs.ToolDir, item+".json")
		var tool Tool
		if err := tool.Init(toolPath); err != nil {
			return fmt.Errorf("cannot read tool: %s", toolPath)
		}

		absToolPath, err := filepath.Abs(tool.RunPath)
		if err != nil {
			return fmt.Errorf("cannot get absolute path of tool path: %s", toolPath)
		}

		toolchain.WriteString(fmt.Sprintf("set(ENV{PATH} \"%s:$ENV{PATH}\")\n", absToolPath))
		environment.WriteString(fmt.Sprintf("export PATH=%s:$PATH\n", absToolPath))

		// Make sure the tool is in the PATH of current process.
		os.Setenv("PATH", fmt.Sprintf("%s%c%s", absToolPath, os.PathListSeparator, os.Getenv("PATH")))
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
