package config

import (
	"buildenv/pkg/io"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type BuildEnvCallbacks interface {
	OnCreatePlatform(platformName string) error
	OnSelectPlatform(platformName string) error
	OnCreateProject(projectName string) error
	OnSelectProject(projectName string) error
	OnCreateTool(toolName string) error
	OnCreatePort(portNameVersion string) error
	About(version string) string
}

type Platform struct {
	RootFS    *RootFS    `json:"rootfs"`
	Toolchain *Toolchain `json:"toolchain"`
	Tools     []string   `json:"tools"`

	// Internal fields.
	Name string  `json:"-"`
	ctx  Context `json:"-"`
}

func (p *Platform) Init(ctx Context, platformName string) error {
	p.ctx = ctx

	// Check if platform name is empty.
	platformName = strings.TrimSpace(platformName)
	if platformName == "" {
		return fmt.Errorf("no platform has been selected for buildenv")
	}

	// Check if platform file exists.
	platformPath := filepath.Join(Dirs.PlatformsDir, platformName+".json")
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
	p.Name = platformName
	return nil
}

func (p Platform) Write(platformPath string) error {
	// Create empty array for empty field.
	p.RootFS = new(RootFS)
	p.Toolchain = new(Toolchain)

	if len(p.Tools) == 0 {
		p.Tools = []string{}
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
		toolpath := filepath.Join(Dirs.ToolsDir, item+".json")
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

		os.Setenv("PATH", absToolPath+string(os.PathListSeparator)+os.Getenv("PATH"))
	}

	// Append $PKG_CONFIG_PATH with pkgconfig path that in installed dir.
	installedDir := filepath.Join(Dirs.WorkspaceDir, "installed", p.Name+"-"+args.BuildType())
	os.Setenv("PKG_CONFIG_PATH", installedDir+"/lib/pkgconfig"+string(os.PathListSeparator)+os.Getenv("PKG_CONFIG_PATH"))

	return nil
}
