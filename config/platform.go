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
	RootFS    *RootFS    `json:"rootfs"`
	Toolchain *Toolchain `json:"toolchain"`
	Tools     []string   `json:"tools"`
	Packages  []string   `json:"packages"`

	// Internal fields.
	platformName string `json:"-"`
}

func (p *Platform) Init(platformPath string) error {
	// Check if platform file exists.
	if !pathExists(platformPath) {
		return fmt.Errorf("platform file not exists: %s", platformPath)
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
	p.platformName = strings.TrimSuffix(filepath.Base(platformPath), ".json")
	return nil
}

func (p Platform) Write(platformPath string) error {
	// Create empty array for empty field.
	p.RootFS = new(RootFS)
	p.Toolchain = new(Toolchain)

	if len(p.RootFS.EnvVars.PKG_CONFIG_PATH) == 0 {
		p.RootFS.EnvVars.PKG_CONFIG_PATH = []string{}
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
	if pathExists(platformPath) {
		return fmt.Errorf("[%s] is already exists", platformPath)
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
		if err := p.RootFS.Verify(args); err != nil {
			return fmt.Errorf("buildenv.rootfs error: %w", err)
		}
	}

	// Toolchain maybe nil when platform is native.
	if p.Toolchain != nil {
		if err := p.Toolchain.Verify(args); err != nil {
			return fmt.Errorf("buildenv.toolchain error: %w", err)
		}
	}

	// Verify tools.
	for _, item := range p.Tools {
		toolpath := filepath.Join(Dirs.ToolDir, item+".json")
		var tool Tool

		if err := tool.Init(toolpath); err != nil {
			return fmt.Errorf("buildenv.tools[%s] read error: %w", item, err)
		}

		if err := tool.Verify(args); err != nil {
			return fmt.Errorf("buildenv.tools[%s] verify error: %w", item, err)
		}
	}

	// Append $PKG_CONFIG_PATH with pkgconfig path that in installed dir.
	installedDir := filepath.Join(Dirs.WorkspaceDir, "installed", p.platformName+"-"+args.BuildType)
	os.Setenv("PKG_CONFIG_PATH", fmt.Sprintf("%s/lib/pkgconfig:%s", installedDir, os.Getenv("PKG_CONFIG_PATH")))

	// Verify dependencies.
	for _, item := range p.Packages {
		portPath := filepath.Join(Dirs.PortDir, item+".json")
		var port Port
		if err := port.Init(portPath, p.platformName, args.BuildType); err != nil {
			return fmt.Errorf("buildenv.packages[%s] read error: %w", item, err)
		}

		if err := port.Verify(args); err != nil {
			return fmt.Errorf("buildenv.packages[%s] verify error: %w", item, err)
		}
	}

	return nil
}

func (p Platform) CreateToolchainFile(scriptDir string) (string, error) {
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
	if p.Toolchain != nil {
		toolchain.WriteString("\n# Set toolchain platform infos.\n")
		toolchain.WriteString(fmt.Sprintf("set(CMAKE_SYSTEM_NAME \"%s\")\n", p.Toolchain.SystemName))
		toolchain.WriteString(fmt.Sprintf("set(CMAKE_SYSTEM_PROCESSOR \"%s\")\n", p.Toolchain.SystemProcessor))
	}

	// Set toolchain for cross-compile.
	if p.Toolchain != nil {
		if err := p.Toolchain.generate(&toolchain, &environment); err != nil {
			return "", err
		}
	}

	// Set sysroot for cross-compile.
	if p.RootFS != nil {
		if err := p.RootFS.generate(&toolchain, &environment); err != nil {
			return "", err
		}
	}

	// Set tools for cross-compile.
	if err := p.writeTools(&toolchain, &environment); err != nil {
		return "", err
	}

	toolchain.WriteString("\n# Add `installed dir` into library search paths.\n")
	installedDir := filepath.Join(Dirs.WorkspaceDir, "installed", p.platformName+"-${CMAKE_BUILD_TYPE}")
	toolchain.WriteString(fmt.Sprintf("list(APPEND CMAKE_FIND_ROOT_PATH \"%s\")\n", installedDir))
	toolchain.WriteString(fmt.Sprintf("list(APPEND CMAKE_PREFIX_PATH \"%s\")\n", installedDir))
	toolchain.WriteString(fmt.Sprintf("list(APPEND ENV{PKG_CONFIG_PATH} \"%s/lib/pkgconfig\")\n", installedDir))

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

func (b *Platform) writeTools(toolchain, environment *strings.Builder) error {
	toolchain.WriteString("\n# Append `path` of tools into $PATH.\n")
	environment.WriteString("\n# Append `path` of tools into $PATH.\n")

	for _, item := range b.Tools {
		toolPath := filepath.Join(Dirs.ToolDir, item+".json")
		var tool Tool
		if err := tool.Init(toolPath); err != nil {
			return fmt.Errorf("cannot read tool: %s", toolPath)
		}

		absToolPath, err := filepath.Abs(tool.Path)
		if err != nil {
			return fmt.Errorf("cannot get absolute path of tool path: %s", toolPath)
		}

		toolchain.WriteString(fmt.Sprintf("list(APPEND ENV{PATH} \"%s\")\n", absToolPath))
		environment.WriteString(fmt.Sprintf("export PATH=%s:$PATH\n", absToolPath))

		// Append $PATH with tool path.
		os.Setenv("PATH", fmt.Sprintf("%s%c%s", absToolPath, os.PathListSeparator, os.Getenv("PATH")))
	}
	return nil
}
