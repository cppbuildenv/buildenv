package buildsystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func NewMake(config BuildConfig) *make {
	return &make{BuildConfig: config}
}

type make struct {
	BuildConfig
}

func (m make) Configure(buildType string) error {
	// Remove build dir and create it for configure.
	if err := os.RemoveAll(m.PortConfig.BuildDir); err != nil {
		return err
	}
	if err := os.MkdirAll(m.PortConfig.BuildDir, os.ModeDir|os.ModePerm); err != nil {
		return err
	}

	if err := os.Chdir(m.PortConfig.BuildDir); err != nil {
		return err
	}

	// Append common variables for cross compiling.
	m.Arguments = append(m.Arguments, fmt.Sprintf("--prefix=%s", m.PortConfig.PackageDir))

	// Replace placeholders with real paths.
	for index, argument := range m.Arguments {
		if strings.Contains(argument, "${HOST}") {
			m.Arguments[index] = strings.ReplaceAll(argument, "${HOST}", m.PortConfig.Host)
		}

		if strings.Contains(argument, "${SYSTEM_NAME}") {
			m.Arguments[index] = strings.ReplaceAll(argument, "${SYSTEM_NAME}", strings.ToLower(m.PortConfig.SystemName))
		}

		if strings.Contains(argument, "${SYSTEM_PROCESSOR}") {
			m.Arguments[index] = strings.ReplaceAll(argument, "${SYSTEM_PROCESSOR}", m.PortConfig.SystemProcessor)
		}

		if strings.Contains(argument, "${SYSROOT}") {
			m.Arguments[index] = strings.ReplaceAll(argument, "${SYSROOT}", m.PortConfig.RootFS)
		}

		if strings.Contains(argument, "${CROSS_PREFIX}") {
			m.Arguments[index] = strings.ReplaceAll(argument, "${CROSS_PREFIX}", m.PortConfig.ToolchainPrefix)
		}
	}

	// Join args into a string.
	joinedArgs := strings.Join(m.Arguments, " ")
	configure := fmt.Sprintf("%s/configure %s", m.PortConfig.SourceDir, joinedArgs)

	// Execute configure.
	parentDir := filepath.Dir(m.PortConfig.BuildDir)
	fileName := filepath.Base(m.PortConfig.BuildDir) + "-configure.log"
	configureLogPath := filepath.Join(parentDir, fileName)
	title := fmt.Sprintf("[configure %s]", m.PortConfig.LibName)
	if err := m.execute(title, configure, configureLogPath); err != nil {
		return err
	}

	return nil
}

func (m make) Build() error {
	// Assemble script.
	command := fmt.Sprintf("make -j %d", m.PortConfig.JobNum)

	// Execute build.
	parentDir := filepath.Dir(m.PortConfig.BuildDir)
	fileName := filepath.Base(m.PortConfig.BuildDir) + "-build.log"
	buildLogPath := filepath.Join(parentDir, fileName)
	title := fmt.Sprintf("[build %s]", m.PortConfig.LibName)
	if err := m.execute(title, command, buildLogPath); err != nil {
		return err
	}

	return nil
}

func (m make) Install() error {
	// Assemble script.
	command := "make install"

	// Execute install.
	parentDir := filepath.Dir(m.PortConfig.BuildDir)
	fileName := filepath.Base(m.PortConfig.BuildDir) + "-install.log"
	installLogPath := filepath.Join(parentDir, fileName)
	title := fmt.Sprintf("[install %s]", m.PortConfig.LibName)
	if err := m.execute(title, command, installLogPath); err != nil {
		return err
	}

	return nil
}
