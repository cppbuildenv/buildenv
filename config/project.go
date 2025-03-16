package config

import (
	"buildenv/buildsystem"
	"buildenv/pkg/fileio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type Project struct {
	Ports         []string                           `json:"ports"`
	OverridePorts map[string]buildsystem.BuildConfig `json:"override_ports"`
	CMakeVars     []string                           `json:"cmake_vars"`
	EnvVars       []string                           `json:"env_vars"`
	MicroVars     []string                           `json:"micro_vars"`

	// Internal fields.
	Name          string                    `json:"-"`
	ctx           Context                   `json:"-"`
	trackingPorts map[string][]trackingInfo `json:"-"`
}

type trackingInfo struct {
	version string
	parent  string
}

func (p *Project) Init(ctx Context, projectName string) error {
	p.ctx = ctx

	// Check if project name is empty.
	projectName = strings.TrimSpace(projectName)
	if projectName == "" {
		return fmt.Errorf("no project has been selected for buildenv")
	}

	// Check if project file exists.
	projectPath := filepath.Join(Dirs.ProjectsDir, projectName+".json")
	if !fileio.PathExists(projectPath) {
		return fmt.Errorf("project %s does not exists", projectName)
	}

	// Read conf/buildenv.json
	bytes, err := os.ReadFile(projectPath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bytes, p); err != nil {
		return fmt.Errorf("read error: %w", err)
	}

	// Set values of internal fields.
	p.Name = projectName
	return nil
}

func (p Project) Write(platformPath string) error {
	if len(p.Ports) == 0 {
		p.Ports = []string{}
	}

	bytes, err := json.MarshalIndent(p, "", "    ")
	if err != nil {
		return err
	}

	// Check if conf/buildenv.json exists.
	if fileio.PathExists(platformPath) {
		return fmt.Errorf("%s is already exists", platformPath)
	}

	// Make sure the parent directory exists.
	parentDir := filepath.Dir(platformPath)
	if err := os.MkdirAll(parentDir, os.ModeDir|os.ModePerm); err != nil {
		return err
	}

	return os.WriteFile(platformPath, bytes, os.ModePerm)
}

func (p Project) Setup(args SetupArgs) error {
	// Check if ports version conflicts in the project.
	if err := p.checkPortsConflicts(); err != nil {
		return err
	}

	// Validate dependencies.
	validatePort := func(nameVersion string) error {
		portPath := filepath.Join(Dirs.PortsDir, nameVersion+".json")
		var port Port
		if err := port.Init(p.ctx, portPath); err != nil {
			return fmt.Errorf("%s: %w", nameVersion, err)
		}

		if err := port.Validate(); err != nil {
			return fmt.Errorf("%s: %w", nameVersion, err)
		}

		if args.InstallPorts() {
			if err := port.Install(args.Silent()); err != nil {
				return fmt.Errorf("%s: %w", nameVersion, err)
			}
		}

		return nil
	}

	for _, item := range p.Ports {
		if err := validatePort(item); err != nil {
			return err
		}
	}

	return nil
}

func (p *Project) checkPortsConflicts() error {
	p.trackingPorts = make(map[string][]trackingInfo)

	for _, nameVersion := range p.Ports {
		var port Port
		if err := port.Init(p.ctx, filepath.Join(Dirs.PortsDir, nameVersion+".json")); err != nil {
			return err
		}

		p.trackingPorts[port.Name] = []trackingInfo{
			{
				version: port.Version,
				parent:  p.Name,
			},
		}

		if err := p.trackingPortDepedencies(port); err != nil {
			return err
		}
	}

	// Check if there are any conflicts.
	var summaries []string
	for portName, trackingInfos := range p.trackingPorts {
		if len(trackingInfos) > 1 {
			var conflicts []string
			for _, trackingInfo := range trackingInfos {
				conflicts = append(conflicts, fmt.Sprintf("%s@%s is defined in %s", portName, trackingInfo.version, trackingInfo.parent))
			}

			summaries = append(summaries, fmt.Sprintf("    - %s", strings.Join(conflicts, ", ")))
		}
	}
	if len(summaries) > 0 {
		return fmt.Errorf("detected conflicting versions of ports:\n%s", strings.Join(summaries, "\n"))
	}

	return nil
}

func (p *Project) trackingPortDepedencies(port Port) error {
	// Find matched config and init build system.
	var matchedConfig *buildsystem.BuildConfig
	for _, config := range port.BuildConfigs {
		if port.MatchPattern(config.Pattern) {
			if err := config.InitBuildSystem(); err != nil {
				return err
			}
			matchedConfig = &config
			break
		}
	}
	if matchedConfig == nil {
		return fmt.Errorf("no matching build_config found to build for %s", port.NameVersion())
	}

	// Tracking port depedencies infos.
	for _, depedency := range matchedConfig.Depedencies {
		var subPort Port
		if err := subPort.Init(p.ctx, filepath.Join(Dirs.PortsDir, depedency+".json")); err != nil {
			return err
		}

		if infos, ok := p.trackingPorts[subPort.Name]; ok {
			contains := slices.ContainsFunc(infos, func(info trackingInfo) bool {
				return info.version == subPort.Version
			})
			if !contains {
				p.trackingPorts[subPort.Name] = append(p.trackingPorts[subPort.Name], trackingInfo{
					version: subPort.Version,
					parent:  port.Name,
				})
			}
		} else {
			p.trackingPorts[subPort.Name] = append(p.trackingPorts[subPort.Name], trackingInfo{
				version: subPort.Version,
				parent:  port.Name,
			})
		}
	}

	return nil
}
