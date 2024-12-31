package config

import (
	"buildenv/pkg/io"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Project struct {
	Ports     []string `json:"ports"`
	CMakeVars []string `json:"cmake_vars"`
	EnvVars   []string `json:"env_vars"`
	MicroVars []string `json:"micro_vars"`

	// Internal fields.
	Name string  `json:"-"`
	ctx  Context `json:"-"`
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
	if !io.PathExists(projectPath) {
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

func (p Project) Verify(request VerifyRequest) error {
	verifyPort := func(portNameVersion string) error {
		portPath := filepath.Join(Dirs.PortsDir, portNameVersion+".json")
		var port Port
		if err := port.Init(p.ctx, portPath); err != nil {
			return fmt.Errorf("%s: %w", portNameVersion, err)
		}

		if err := port.Verify(); err != nil {
			return fmt.Errorf("%s: %w", portNameVersion, err)
		}

		if request.InstallPorts() {
			if err := port.Install(request.Silent()); err != nil {
				return fmt.Errorf("%s: %w", portNameVersion, err)
			}
		}

		return nil
	}

	// Verify dependencies.
	for _, item := range p.Ports {
		if err := verifyPort(item); err != nil {
			return err
		}
	}

	return nil
}
