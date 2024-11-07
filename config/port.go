package config

import (
	"buildenv/config/buildsystem"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type BuildTool int

type Port struct {
	Repo        string                  `json:"repo"`
	Ref         string                  `json:"ref"`
	Depedencies []string                `json:"dependencies"`
	BuildConfig buildsystem.BuildConfig `json:"build_config"`

	// Internal fields.
	portName string `json:"-"`
}

func (p *Port) Read(filePath string) error {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, p); err != nil {
		return err
	}

	portName := strings.TrimSuffix(filepath.Base(p.Repo), ".git") + "-" + p.Ref

	// Set default build dir and installed dir and also can be changed during units tests.
	p.BuildConfig.BuildDir, _ = filepath.Abs(filepath.Join(Dirs.WorkspaceDir, "buildtrees", portName, "x86_64-linux-Release"))
	p.BuildConfig.SourceDir, _ = filepath.Abs(filepath.Join(Dirs.WorkspaceDir, "buildtrees", portName, "src"))
	p.BuildConfig.InstalledDir, _ = filepath.Abs(filepath.Join(Dirs.WorkspaceDir, "installed", "x86_64-linux-Release"))
	p.BuildConfig.JobNum = 8

	p.portName = portName
	return nil
}

func (p *Port) Verify(checkAndRepair bool) error {
	if p.Repo == "" {
		return fmt.Errorf("port.repo is empty")
	}

	if p.Ref == "" {
		return fmt.Errorf("port.ref is empty")
	}

	if p.BuildConfig.BuildTool == "" {
		return fmt.Errorf("port.build_tool is empty")
	}

	if !checkAndRepair {
		return nil
	}

	if err := p.checkAndRepair(); err != nil {
		return err
	}

	return nil
}

func (p Port) Installed() bool {
	return false
}

func (p Port) checkAndRepair() error {
	var buildSystem buildsystem.BuildSystem

	switch p.BuildConfig.BuildTool {
	case "cmake":
		buildSystem = buildsystem.NewCMake(p.BuildConfig)
	case "ninja":
		buildSystem = buildsystem.NewNinja(p.BuildConfig)
	case "make":
		buildSystem = buildsystem.NewMake(p.BuildConfig)
	case "autotools":
		buildSystem = buildsystem.NewAutoTool(p.BuildConfig)
	case "meson":
		buildSystem = buildsystem.NewMeson(p.BuildConfig)
	default:
		return fmt.Errorf("unsupported build system: %s", p.BuildConfig.BuildTool)
	}

	if err := buildSystem.Clone(p.Repo, p.Ref); err != nil {
		return err
	}

	if err := buildSystem.Configure(); err != nil {
		return err
	}

	if err := buildSystem.Build(); err != nil {
		return err
	}

	if err := buildSystem.Install(); err != nil {
		return err
	}

	fmt.Printf("[✔] -------- %s\n\n", p.portName)
	return nil
}
