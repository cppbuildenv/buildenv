package config

import (
	"bufio"
	"buildenv/config/build"
	"buildenv/config/deploy"
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
	Url          string               `json:"url"`
	Version      string               `json:"version"`
	SourceFolder string               `json:"source_folder,omitempty"`
	BuildConfig  *build.BuildConfig   `json:"build_config"`
	DeployConfig *deploy.DeployConfig `json:"deploy_config"`

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

	if p.BuildConfig != nil {
		p.BuildConfig.SourceDir = filepath.Join(Dirs.WorkspaceDir, "buildtrees", portName, "src")
		p.BuildConfig.SourceFolder = p.SourceFolder
		p.BuildConfig.BuildDir = filepath.Join(Dirs.WorkspaceDir, "buildtrees", portName, platformName+"-"+buildType)
		p.BuildConfig.InstalledDir = filepath.Join(Dirs.WorkspaceDir, "installed", platformName+"-"+buildType)
		p.BuildConfig.JobNum = 8 // TODO: make it configurable.
	} else {
		p.DeployConfig = &deploy.DeployConfig{}
		p.DeployConfig.InstalledDir = filepath.Join(Dirs.WorkspaceDir, "installed", platformName+"-"+buildType)
		p.DeployConfig.DownloadDir = filepath.Join(Dirs.WorkspaceDir, "downloads")
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

	if p.BuildConfig != nil {
		if err := p.BuildConfig.Verify(); err != nil {
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

	if p.BuildConfig != nil {
		// Check and repair dependencies.
		for _, item := range p.BuildConfig.Depedencies {
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

		if err := p.BuildConfig.CheckAndRepair(p.Url, p.Version, p.buildType); err != nil {
			return err
		}
	} else {
		if err := p.DeployConfig.CheckAndRepair(p.Url); err != nil {
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
