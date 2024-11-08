package config

import (
	"bufio"
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
	portName     string `json:"-"`
	platformName string `json:"-"`
	buildType    string `json:"-"`
	infoPath     string `json:"-"`
}

func (p *Port) Init(portPath, platformName, buildType string) error {
	bytes, err := os.ReadFile(portPath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, p); err != nil {
		return err
	}

	portName := strings.TrimSuffix(filepath.Base(p.Repo), ".git") + "-" + p.Ref

	// Set default build dir and installed dir and also can be changed during units tests.
	p.BuildConfig.SourceDir = filepath.Join(Dirs.WorkspaceDir, "buildtrees", portName, "src")
	p.BuildConfig.BuildDir = filepath.Join(Dirs.WorkspaceDir, "buildtrees", portName, platformName+"-"+buildType)
	p.BuildConfig.InstalledDir = filepath.Join(Dirs.WorkspaceDir, "installed", platformName+"-"+buildType)
	p.BuildConfig.JobNum = 8 // TODO: make it configurable.

	p.portName = portName
	p.platformName = platformName
	p.buildType = buildType

	// Info file: used to record installed state.
	fileName := fmt.Sprintf("%s-%s.list", p.platformName, p.buildType)
	p.infoPath = filepath.Join(Dirs.InstalledDir, "buildenv", fileName)

	return nil
}

func (p *Port) Verify(args VerifyArgs) error {
	if p.Repo == "" {
		return fmt.Errorf("port.repo is empty")
	}

	if p.Ref == "" {
		return fmt.Errorf("port.ref is empty")
	}

	if p.BuildConfig.BuildTool == "" {
		return fmt.Errorf("port.build_tool is empty")
	}

	if !args.CheckAndRepair {
		return nil
	}

	if err := p.checkAndRepair(); err != nil {
		return err
	}

	return nil
}

func (p Port) Installed() bool {
	if !pathExists(p.infoPath) {
		return false
	}

	// Open the file and read its content.
	file, err := os.OpenFile(p.infoPath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return false
	}
	defer file.Close()

	// Scan through the file line by line to check if the port is installed.
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, p.portName) {
			return true
		}
	}

	return false
}

func (p Port) checkAndRepair() error {
	// No need to check and repair if the port is already installed.
	if p.Installed() {
		return nil
	}

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

	if err := buildSystem.Configure(p.buildType); err != nil {
		return err
	}

	if err := buildSystem.Build(); err != nil {
		return err
	}

	if err := buildSystem.Install(); err != nil {
		return err
	}

	if !pathExists(p.infoPath) {
		if err := os.MkdirAll(filepath.Dir(p.infoPath), os.ModePerm); err != nil {
			return err
		}
	}

	if err := os.WriteFile(p.infoPath, []byte(p.portName+"\n"), os.ModeAppend|os.ModePerm); err != nil {
		return err
	}

	fmt.Printf("[✔] -------- %s (port: %s)\n\n", p.portName, p.BuildConfig.InstalledDir)
	return nil
}
