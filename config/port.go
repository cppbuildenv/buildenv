package config

import (
	"buildenv/buildsystem"
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
	Url          string                    `json:"url"`
	Name         string                    `json:"name"`
	Version      string                    `json:"version"`
	SourceFolder string                    `json:"source_folder,omitempty"`
	BuildConfigs []buildsystem.BuildConfig `json:"build_configs"`

	// Internal fields.
	ctx             Context
	installInfoFile string // used to record installed state
	portsDir        string // it should be `conf/ports`
	isSubDep        bool
}

func (p Port) NameVersion() string {
	p.Version = strings.ReplaceAll(p.Version, "/", "^")
	p.Version = strings.ReplaceAll(p.Version, ":", "^")
	p.Version = strings.ReplaceAll(p.Version, "-", "^")

	return p.Name + "-" + p.Version
}

func (p *Port) Init(ctx Context, portPath string) error {
	if !io.PathExists(portPath) {
		portName := io.FileBaseName(portPath)
		if p.isSubDep {
			return fmt.Errorf("sub depedency port %s does not exists", portName)
		} else {
			return fmt.Errorf("port %s does not exists", portName)
		}
	}

	bytes, err := os.ReadFile(portPath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, p); err != nil {
		return err
	}

	// Info file: used to record installed state.
	nameVersion := p.NameVersion()
	platformBuildType := fmt.Sprintf("%s-%s", ctx.Platform().Name, ctx.BuildType())
	fileName := fmt.Sprintf("%s-%s.list", ctx.Platform().Name, ctx.BuildType())
	sourceDir := filepath.Join(Dirs.WorkspaceDir, "buildtrees", nameVersion, "src")
	buildDir := filepath.Join(Dirs.WorkspaceDir, "buildtrees", nameVersion, platformBuildType)
	installedDir := filepath.Join(Dirs.WorkspaceDir, "installed", platformBuildType)

	p.ctx = ctx
	p.portsDir = filepath.Dir(portPath)
	p.installInfoFile = filepath.Join(Dirs.InstalledRootDir, "buildenv", "info", nameVersion+"-"+fileName)

	if len(p.BuildConfigs) > 0 {
		for index := range p.BuildConfigs {
			p.BuildConfigs[index].Version = p.Version
			p.BuildConfigs[index].SystemName = ctx.SystemName()
			p.BuildConfigs[index].LibName = p.Name
			p.BuildConfigs[index].SourceDir = sourceDir
			p.BuildConfigs[index].SourceFolder = p.SourceFolder
			p.BuildConfigs[index].BuildDir = buildDir
			p.BuildConfigs[index].InstalledDir = installedDir
			p.BuildConfigs[index].InstalledRootDir = Dirs.InstalledRootDir
			p.BuildConfigs[index].JobNum = p.ctx.JobNum()
		}
	}

	return nil
}

func (p *Port) Verify() error {
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
		if !p.MatchPattern(config.Pattern) {
			continue
		}

		if err := config.Verify(); err != nil {
			return err
		}
	}

	return nil
}

func (p Port) Installed() bool {
	if !io.PathExists(p.installInfoFile) {
		return false
	}

	// File can be read?
	bytes, err := os.ReadFile(p.installInfoFile)
	if err != nil {
		return false
	}

	// No content?
	if len(bytes) == 0 {
		return false
	}

	return true
}

func (p Port) CheckAndRepair(args VerifyArgs) error {
	if !args.CheckAndRepair() {
		return nil
	}

	installedDir := filepath.Join(Dirs.WorkspaceDir, "installed", p.ctx.Platform().Name+"-"+p.ctx.BuildType())
	if p.Installed() {
		if !args.Silent() {
			title := color.Sprintf(color.Green, "\n[✔] ---- Port: %s\n", p.NameVersion())
			fmt.Printf("%sLocation: %s\n", title, installedDir)
		}
		return nil
	}

	// No config found, download and deploy it.
	if len(p.BuildConfigs) == 0 {
		downloadedDir := filepath.Join(Dirs.WorkspaceDir, "downloads")
		if err := downloadAndDeploy(p.Url, installedDir, downloadedDir); err != nil {
			return err
		}
	}

	// Find matched config.
	var matchedConfig *buildsystem.BuildConfig
	for _, config := range p.BuildConfigs {
		if p.MatchPattern(config.Pattern) {
			matchedConfig = &config
			break
		}
	}
	if matchedConfig == nil {
		return fmt.Errorf("no matching build_config found to build")
	}

	// First, we must check and repair dependency ports.
	for _, item := range matchedConfig.Depedencies {
		if strings.HasPrefix(item, p.Name) {
			return fmt.Errorf("port.dependencies contains circular dependency: %s", item)
		}

		// Check and repair dependency.
		var port Port
		port.isSubDep = true
		portPath := filepath.Join(p.portsDir, item+".json")
		if err := port.Init(p.ctx, portPath); err != nil {
			return err
		}
		if err := port.Verify(); err != nil {
			return err
		}
		if err := port.CheckAndRepair(args); err != nil {
			return err
		}
	}

	// Check and repair current port.
	installLogPath, err := matchedConfig.CheckAndRepair(p.Url, p.Version, p.ctx.BuildType(), matchedConfig.CMakeConfig)
	if err != nil {
		return err
	}

	// Mkdir if not exists.
	if err := os.MkdirAll(filepath.Dir(p.installInfoFile), os.ModePerm); err != nil {
		return err
	}

	// Write installed file info list.
	installedFiles, err := matchedConfig.BuildSystem.InstalledFiles(installLogPath)
	if err != nil {
		return err
	}
	os.WriteFile(p.installInfoFile, []byte(strings.Join(installedFiles, "\n")), os.ModePerm)

	if !args.Silent() {
		title := color.Sprintf(color.Green, "\n[✔] ---- Port: %s\n", p.NameVersion())
		fmt.Printf("%sLocation: %s\n", title, installedDir)
	}

	return nil
}

func downloadAndDeploy(url, installedDir, downloadedDir string) error {
	// Download to fixed dir.
	downloadRequest := io.NewDownloadRequest(url, downloadedDir)
	downloaded, err := downloadRequest.Download()
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

func (p Port) MatchPattern(pattern string) bool {
	pattern = strings.TrimSpace(pattern)

	if pattern == "" || pattern == "*" {
		return true
	}

	platformName := p.ctx.Platform().Name
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
