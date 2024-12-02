package config

import (
	"buildenv/pkg/env"
	"buildenv/pkg/io"
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
	SetOffline(offline bool) error
}

type Platform struct {
	RootFS    *RootFS    `json:"rootfs"`
	Toolchain *Toolchain `json:"toolchain"`
	Tools     []string   `json:"tools"`
	Packages  []string   `json:"packages"`

	// Internal fields.
	platformName string
	ctx          Context
}

func (p *Platform) Init(ctx Context, platformName string) error {
	p.ctx = ctx

	// Check if platform name is empty.
	platformName = strings.TrimSpace(platformName)
	if platformName == "" {
		return fmt.Errorf("no platform has been selected for buildenv")
	}

	// Check if platform file exists.
	platformPath := filepath.Join(Dirs.PlatformDir, platformName+".json")
	if !io.PathExists(platformPath) {
		return fmt.Errorf("platform %s does not exists", platformName)
	}

	// Read conf/buildenv.json
	bytes, err := os.ReadFile(platformPath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bytes, p); err != nil {
		return fmt.Errorf("read error: %w", err)
	}

	// Set values of internal fields.
	p.platformName = platformName
	return nil
}

func (p Platform) Write(platformPath string) error {
	// Create empty array for empty field.
	p.RootFS = new(RootFS)
	p.Toolchain = new(Toolchain)

	if len(p.RootFS.PkgConfigPath) == 0 {
		p.RootFS.PkgConfigPath = []string{}
	}
	if len(p.Tools) == 0 {
		p.Tools = []string{}
	}
	if len(p.Packages) == 0 {
		p.Packages = []string{}
	}

	bytes, err := json.MarshalIndent(p, "", "    ")
	if err != nil {
		return err
	}

	// Check if conf/buildenv.json exists.
	if io.PathExists(platformPath) {
		return fmt.Errorf("%s is already exists", platformPath)
	}

	// Make sure the parent directory exists.
	parentDir := filepath.Dir(platformPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(platformPath, bytes, os.ModePerm)
}

func (p Platform) Verify(args VerifyArgs) error {
	// RootFS maybe nil when platform is native.
	if p.RootFS != nil {
		if err := p.RootFS.Verify(); err != nil {
			return err
		}

		if err := p.RootFS.CheckAndRepair(args); err != nil {
			return fmt.Errorf("buildenv.rootfs check and repair error: %w", err)
		}
	}

	// Toolchain maybe nil when platform is native.
	if p.Toolchain != nil {
		if err := p.Toolchain.Verify(); err != nil {
			return fmt.Errorf("buildenv.toolchain error: %w", err)
		}

		if err := p.Toolchain.CheckAndRepair(args); err != nil {
			return fmt.Errorf("buildenv.toolchain check and repair error: %w", err)
		}
	}

	// Verify tools.
	for _, item := range p.Tools {
		toolpath := filepath.Join(Dirs.ToolDir, item+".json")
		var tool Tool

		if err := tool.Init(toolpath); err != nil {
			return fmt.Errorf("buildenv.tools[%s] read error: %w", item, err)
		}

		if err := tool.Verify(); err != nil {
			return fmt.Errorf("buildenv.tools[%s] verify error: %w", item, err)
		}

		if err := tool.CheckAndRepair(args); err != nil {
			return fmt.Errorf("buildenv.tools[%s] check and repair error: %w", item, err)
		}

		// Append $PATH with tool path.
		absToolPath, err := filepath.Abs(tool.Path)
		if err != nil {
			return fmt.Errorf("cannot get absolute path of tool path: %s", tool.Path)
		}

		os.Setenv("PATH", fmt.Sprintf("%s%c%s", absToolPath, os.PathListSeparator, os.Getenv("PATH")))
	}

	// Append $PKG_CONFIG_PATH with pkgconfig path that in installed dir.
	installedDir := filepath.Join(Dirs.WorkspaceDir, "installed", p.platformName+"-"+args.BuildType())
	os.Setenv("PKG_CONFIG_PATH", fmt.Sprintf("%s/lib/pkgconfig:%s", installedDir, os.Getenv("PKG_CONFIG_PATH")))

	// Inner function used to verify port.
	verifyPort := func(portName string) error {
		portPath := filepath.Join(Dirs.PortDir, portName+".json")
		var port Port
		if err := port.Init(p.ctx, portPath); err != nil {
			return fmt.Errorf("buildenv.packages[%s] read error: %w", portName, err)
		}

		if err := port.Verify(); err != nil {
			return fmt.Errorf("buildenv.packages[%s] verify error: %w", portName, err)
		}

		if err := port.CheckAndRepair(args); err != nil {
			return fmt.Errorf("buildenv.packages[%s] check and repair error: %w", portName, err)
		}

		return nil
	}

	// Check if only to verify one port.
	portNeedToVerify := args.PackagePort()
	if portNeedToVerify != "" {
		verifyPort(portNeedToVerify)
	} else {
		// Verify dependencies.
		for _, item := range p.Packages {
			verifyPort(item)
		}
	}

	return nil
}

func (p Platform) GenerateToolchainFile(scriptDir string) (string, error) {
	var toolchain, environment strings.Builder

	// Verify buildenv during configuration.
	toolchain.WriteString(fmt.Sprintf("%s\n", `# Set default CMAKE_BUILD_TYPE.
if(NOT CMAKE_BUILD_TYPE)
	set(CMAKE_BUILD_TYPE "Release")
endif()

# Verify buildenv during configuration.
set(HOME_DIR "${CMAKE_CURRENT_LIST_DIR}/..")
find_program(BUILDENV buildenv PATHS ${HOME_DIR})
if(BUILDENV)
	message("================ buildenv -verify -silent -build_type ${CMAKE_BUILD_TYPE} ================\n")
	execute_process(
		COMMAND ${BUILDENV} -verify -silent -build_type=${CMAKE_BUILD_TYPE}
		WORKING_DIRECTORY ${HOME_DIR}
	)
endif()`))

	// Define buildenv root dir.
	toolchain.WriteString(fmt.Sprintf("\n%s\n", `# Define buildenv root dir.
get_filename_component(_CURRENT_DIR "${CMAKE_CURRENT_LIST_FILE}" PATH)
get_filename_component(BUILDENV_ROOT_DIR "${_CURRENT_DIR}" PATH)`))

	environment.WriteString("\n# Define buildenv root dir.\n")
	environment.WriteString("export BUILDENV_ROOT_DIR=$PWD/..\n")

	// Set sysroot for cross-compile.
	if p.RootFS != nil {
		if err := p.RootFS.generate(&toolchain, &environment); err != nil {
			return "", err
		}
	}

	// Set toolchain for cross-compile.
	if p.Toolchain != nil {
		// Set toolchain platform infos.
		toolchain.WriteString("\n# Set toolchain platform infos.\n")
		toolchain.WriteString(fmt.Sprintf("set(CMAKE_SYSTEM_NAME \"%s\")\n", p.Toolchain.SystemName))
		toolchain.WriteString(fmt.Sprintf("set(CMAKE_SYSTEM_PROCESSOR \"%s\")\n", p.Toolchain.SystemProcessor))

		if err := p.Toolchain.generate(&toolchain, &environment); err != nil {
			return "", err
		}
	}

	// Set tools for cross-compile.
	if err := p.writeTools(&toolchain, &environment); err != nil {
		return "", err
	}

	toolchain.WriteString("\n# Add `installed dir` into library search paths.\n")
	installedDir := fmt.Sprintf("${BUILDENV_ROOT_DIR}/%s/%s", "installed", p.platformName+"-${CMAKE_BUILD_TYPE}")
	toolchain.WriteString(fmt.Sprintf("list(APPEND CMAKE_FIND_ROOT_PATH \"%s\")\n", installedDir))
	toolchain.WriteString(fmt.Sprintf("list(APPEND CMAKE_PREFIX_PATH \"%s\")\n", installedDir))
	toolchain.WriteString(fmt.Sprintf("set(ENV{PKG_CONFIG_PATH} \"%s/lib/pkgconfig%s$ENV{PKG_CONFIG_PATH}\")\n", installedDir, string(os.PathListSeparator)))

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

func (p *Platform) writeTools(toolchain, environment *strings.Builder) error {
	toolchain.WriteString("\n# Append `path` of tools into $PATH.\n")
	environment.WriteString("\n# Append `path` of tools into $PATH.\n")

	for _, item := range p.Tools {
		toolPath := filepath.Join(Dirs.ToolDir, item+".json")
		var tool Tool
		if err := tool.Init(toolPath); err != nil {
			return fmt.Errorf("cannot read tool: %s", toolPath)
		}

		if err := tool.Verify(); err != nil {
			return fmt.Errorf("cannot verify tool: %s", toolPath)
		}

		toolchain.WriteString(fmt.Sprintf("set(ENV{PATH} \"%s\")\n", env.Join(tool.cmakepath, "$ENV{PATH}")))
		environment.WriteString(fmt.Sprintf("export PATH=%s\n", env.Join(tool.cmakepath, "$PATH")))
	}
	return nil
}
