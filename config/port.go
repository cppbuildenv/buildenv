package config

import (
	"bufio"
	"buildenv/buildsystem"
	"buildenv/generator"
	"buildenv/pkg/color"
	"buildenv/pkg/io"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type BuildTool int

type Port struct {
	Url                 string                     `json:"url"`
	Name                string                     `json:"name"`
	Version             string                     `json:"version"`
	SourceFolder        string                     `json:"source_folder,omitempty"`
	Depedencies         []string                   `json:"dependencies"`
	BuildConfigs        []buildsystem.BuildConfig  `json:"build_configs"`
	GenerateCMakeConfig *generator.GeneratorConfig `json:"generate_cmake_config"`

	// Internal fields.
	ctx      Context
	fullName string // portName = portName + "-" + p.Version
	infoPath string // used to record installed state
	portDir  string // it should be `conf/ports`
}

func (p *Port) Init(ctx Context, portPath string) error {
	bytes, err := os.ReadFile(portPath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, p); err != nil {
		return err
	}

	// Info file: used to record installed state.
	portNameType := fmt.Sprintf("%s-%s", ctx.Platform(), ctx.BuildType())
	fileName := fmt.Sprintf("%s-%s.list", ctx.Platform(), ctx.BuildType())
	sourceDir := filepath.Join(Dirs.WorkspaceDir, "buildtrees", p.Name, "src")
	buildDir := filepath.Join(Dirs.WorkspaceDir, "buildtrees", p.Name, portNameType)
	installedDir := filepath.Join(Dirs.WorkspaceDir, "installed", portNameType)

	p.ctx = ctx
	p.fullName = p.Name + "-" + p.Version
	p.portDir = filepath.Dir(portPath)
	p.infoPath = filepath.Join(Dirs.InstalledRootDir, fileName)

	if len(p.BuildConfigs) > 0 {
		for index := range p.BuildConfigs {
			p.BuildConfigs[index].Version = p.Version
			p.BuildConfigs[index].SystemName = ctx.SystemName()
			p.BuildConfigs[index].LibName = p.Name
			p.BuildConfigs[index].SourceDir = sourceDir
			p.BuildConfigs[index].SourceFolder = p.SourceFolder
			p.BuildConfigs[index].BuildDir = buildDir
			p.BuildConfigs[index].InstalledDir = installedDir
			p.BuildConfigs[index].JobNum = p.ctx.JobNum()
		}
	}

	return nil
}

func (p *Port) Verify(args VerifyArgs) error {
	if p.Url == "" {
		return fmt.Errorf("port.url is empty")
	}

	if p.Name == "" {
		return fmt.Errorf("port.name is empty")
	}

	if p.Version == "" {
		return fmt.Errorf("port.version is empty")
	}

	for _, config := range p.BuildConfigs {
		if !p.matchPattern(config.Pattern) {
			continue
		}

		if err := config.Verify(); err != nil {
			return err
		}
	}

	if !args.CheckAndRepair() {
		return nil
	}

	if err := p.checkAndRepair(); err != nil {
		return err
	}

	return nil
}

func (p Port) Installed() bool {
	if !io.PathExists(p.infoPath) {
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

		if strings.Contains(line, p.fullName) {
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

	if len(p.BuildConfigs) > 0 {
		// Check and repair dependencies.
		for _, item := range p.Depedencies {
			if item == p.fullName {
				return fmt.Errorf("port.dependencies contains circular dependency: %s", item)
			}

			portPath := filepath.Join(p.portDir, item+".json")

			var port Port
			if err := port.Init(p.ctx, portPath); err != nil {
				return err
			}

			if err := port.checkAndRepair(); err != nil {
				return err
			}
		}

		var matchedAndFixed bool
		for _, config := range p.BuildConfigs {
			if !p.matchPattern(config.Pattern) {
				continue
			}

			if err := config.CheckAndRepair(p.Url, p.Version, p.ctx.BuildType(), p.GenerateCMakeConfig); err != nil {
				return err
			}

			matchedAndFixed = true
		}

		if !matchedAndFixed {
			return fmt.Errorf("no matching build_config found to build")
		}
	} else {
		platformName := p.ctx.Platform()
		buildType := p.ctx.BuildType()

		installedDir := filepath.Join(Dirs.WorkspaceDir, "installed", platformName+"-"+buildType)
		downloadedDir := filepath.Join(Dirs.WorkspaceDir, "downloads")
		if err := downloadAndDeploy(p.Url, installedDir, downloadedDir); err != nil {
			return err
		}
	}

	// Mkdir if not exists.
	if err := os.MkdirAll(filepath.Dir(p.infoPath), os.ModePerm); err != nil {
		return err
	}

	// Write info list file in append mode.
	file, err := os.OpenFile(p.infoPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	if _, err := file.Write([]byte("\n" + p.fullName)); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	platformName := p.ctx.Platform()
	buildType := p.ctx.BuildType()

	installedDir := filepath.Join(Dirs.WorkspaceDir, "installed", platformName+"-"+buildType)
	fmt.Print(color.Sprintf(color.Blue, "[✔] -------- %s (port: %s)\n\n", p.fullName, installedDir))
	return nil
}

func downloadAndDeploy(url, installedDir, downloadedDir string) error {
	// Download to fixed dir.
	downloaded, err := io.Download(url, downloadedDir, "")
	if err != nil {
		return fmt.Errorf("%s: download port failed: %w", url, err)
	}

	// Extract archive file.
	archiveName := filepath.Base(url)
	folderName := strings.TrimSuffix(archiveName, ".tar.gz")
	extractPath := filepath.Join(installedDir, folderName)
	if err := io.Extract(downloaded, extractPath); err != nil {
		return fmt.Errorf("%s: extract %s failed: %w", archiveName, downloaded, err)
	}

	return nil
}

func (p Port) matchPattern(pattern string) bool {
	pattern = strings.TrimSpace(pattern)

	if pattern == "" || pattern == "*" {
		return true
	}

	platformName := p.ctx.Platform()
	if pattern[0] == '*' && pattern[len(pattern)-1] == '*' {
		return strings.Contains(platformName, pattern[1:len(pattern)-1])
	}

	if pattern[0] == '*' {
		return strings.HasSuffix(platformName, pattern[1:])
	}

	if pattern[len(pattern)-1] == '*' {
		return strings.HasPrefix(platformName, pattern[:len(pattern)-1])
	}

	return platformName == pattern
}
