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
	Ports []string `json:"ports"`

	// Internal fields.
	ctx         Context
	projectName string
}

func (p *Project) Init(ctx Context, projectName string) error {
	p.ctx = ctx

	// Check if project name is empty.
	projectName = strings.TrimSpace(projectName)
	if projectName == "" {
		return fmt.Errorf("no project has been selected for buildenv")
	}

	// Check if project file exists.
	projectPath := filepath.Join(Dirs.ProjectDir, projectName+".json")
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
	p.projectName = projectName
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

func (p Project) Verify(args VerifyArgs) error {
	installPort := func(portDesc string) error {
		portPath := filepath.Join(Dirs.PortDir, portDesc+".json")
		var port Port
		if err := port.Init(p.ctx, portPath); err != nil {
			return fmt.Errorf("%s: %w", portDesc, err)
		}

		if err := port.Verify(); err != nil {
			return fmt.Errorf("%s: %w", portDesc, err)
		}

		if err := port.CheckAndRepair(args); err != nil {
			return fmt.Errorf("%s: %w", portDesc, err)
		}

		return nil
	}

	// Check if only to verify one port.
	portToInstall := args.PortToInstall()
	if portToInstall != "" {
		if err := installPort(portToInstall); err != nil {
			return err
		}
	} else {
		// Verify dependencies.
		for _, item := range p.Ports {
			if err := installPort(item); err != nil {
				return err
			}
		}
	}

	return nil
}
