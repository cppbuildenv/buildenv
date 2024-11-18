package config

import (
	"bufio"
	"buildenv/config/buildsystem"
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
	Version      string                    `json:"version"`
	SourceFolder string                    `json:"source_folder,omitempty"`
	Depedencies  []string                  `json:"dependencies"`
	BuildConfigs []buildsystem.BuildConfig `json:"build_configs"`

	// Internal fields.
	portName     string `json:"-"`
	platformName string `json:"-"`
	buildType    string `json:"-"`
	infoPath     string `json:"-"`
	portDir      string `json:"-"`
}

func (p *Port) Init(portPath, platformName, buildType string) error {
	bytes, err := os.ReadFile(portPath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, p); err != nil {
		return err
	}

	portName := strings.TrimSuffix(filepath.Base(p.Url), ".git")
	portName = strings.TrimSuffix(portName, ".tar.gz")
	portName = strings.TrimSuffix(portName, ".tar.xz")

	p.portName = portName + "-" + p.Version
	p.platformName = platformName
	p.buildType = buildType
	p.portDir = filepath.Dir(portPath)

	// Info file: used to record installed state.
	fileName := fmt.Sprintf("%s-%s.list", p.platformName, p.buildType)
	p.infoPath = filepath.Join(Dirs.InstalledRootDir, "buildenv", fileName)

	if len(p.BuildConfigs) > 0 {
		for index := range p.BuildConfigs {
			p.BuildConfigs[index].SourceDir = filepath.Join(Dirs.WorkspaceDir, "buildtrees", portName, "src")
			p.BuildConfigs[index].SourceFolder = p.SourceFolder
			p.BuildConfigs[index].BuildDir = filepath.Join(Dirs.WorkspaceDir, "buildtrees", portName, platformName+"-"+buildType)
			p.BuildConfigs[index].InstalledDir = filepath.Join(Dirs.WorkspaceDir, "installed", platformName+"-"+buildType)
			p.BuildConfigs[index].JobNum = 8 // TODO: make it configurable.
		}
	}

	return nil
}

func (p *Port) Verify(args VerifyArgs) error {
	if p.Url == "" {
		return fmt.Errorf("port.url is empty")
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

	if !args.CheckAndRepair {
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

	if len(p.BuildConfigs) > 0 {
		// Check and repair dependencies.
		for _, item := range p.Depedencies {
			if item == p.portName {
				return fmt.Errorf("port.dependencies contains circular dependency: %s", item)
			}

			portPath := filepath.Join(p.portDir, item+".json")

			var port Port
			if err := port.Init(portPath, p.platformName, p.buildType); err != nil {
				return err
			}

			if err := port.checkAndRepair(); err != nil {
				return err
			}
		}

		for _, config := range p.BuildConfigs {
			if !p.matchPattern(config.Pattern) {
				continue
			}

			if err := config.CheckAndRepair(p.Url, p.Version, p.buildType); err != nil {
				return err
			}
		}
	} else {
		installedDir := filepath.Join(Dirs.WorkspaceDir, "installed", p.platformName+"-"+p.buildType)
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
	if _, err := file.Write([]byte(p.portName + "\n")); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	installedDir := filepath.Join(Dirs.WorkspaceDir, "installed", p.platformName+"-"+p.buildType)
	fmt.Print(color.Sprintf(color.Blue, "[✔] -------- %s (port: %s)\n\n", p.portName, installedDir))
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

	if pattern[0] == '*' && pattern[len(pattern)-1] == '*' {
		return strings.Contains(p.platformName, pattern[1:len(pattern)-1])
	}

	if pattern[0] == '*' {
		return strings.HasSuffix(p.platformName, pattern[1:])
	}

	if pattern[len(pattern)-1] == '*' {
		return strings.HasPrefix(p.platformName, pattern[:len(pattern)-1])
	}

	return p.platformName == pattern
}
