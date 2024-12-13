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
	if err := os.RemoveAll(m.BuildDir); err != nil {
		return err
	}
	if err := os.MkdirAll(m.BuildDir, os.ModeDir|os.ModePerm); err != nil {
		return err
	}

	if err := os.Chdir(m.BuildDir); err != nil {
		return err
	}

	var (
		crossPrefix = os.Getenv("CROSS_PREFIX")
		sysroot     = os.Getenv("SYSROOT")
		host        = os.Getenv("HOST")
	)

	// Append common variables for cross compiling.
	m.Arguments = append(m.Arguments, fmt.Sprintf("--prefix=%s", m.InstalledDir))
	m.Arguments = append(m.Arguments, fmt.Sprintf("--sysroot=%s", sysroot))
	m.Arguments = append(m.Arguments, fmt.Sprintf("--cross-prefix=%s", crossPrefix))

	// Replace placeholders with real paths.
	for index, argument := range m.Arguments {
		if strings.Contains(argument, "${INSTALLED_DIR}") {
			m.Arguments[index] = strings.ReplaceAll(argument, "${INSTALLED_DIR}", m.InstalledDir)
		}

		if strings.Contains(argument, "${HOST}") {
			m.Arguments[index] = strings.ReplaceAll(argument, "${HOST}", host)
		}
	}

	// Join args into a string.
	joinedArgs := strings.Join(m.Arguments, " ")
	configure := fmt.Sprintf("%s/configure %s", m.SourceDir, joinedArgs)

	// Execute configure.
	configureLogPath := filepath.Join(filepath.Dir(m.BuildDir), filepath.Base(m.BuildDir)+"-configure.log")
	title := fmt.Sprintf("[configure %s]", m.LibName)
	if err := m.execute(title, configure, configureLogPath); err != nil {
		return err
	}

	return nil
}

func (m make) Build() error {
	// Assemble script.
	command := fmt.Sprintf("make -j %d", m.JobNum)

	// Execute build.
	buildLogPath := filepath.Join(filepath.Dir(m.BuildDir), filepath.Base(m.BuildDir)+"-build.log")
	title := fmt.Sprintf("[build %s]", m.LibName)
	if err := m.execute(title, command, buildLogPath); err != nil {
		return err
	}

	return nil
}

func (m make) Install() error {
	// Assemble script.
	command := "make install"

	// Execute install.
	installLogPath := filepath.Join(filepath.Dir(m.BuildDir), filepath.Base(m.BuildDir)+"-install.log")
	title := fmt.Sprintf("[install %s]", m.LibName)
	if err := m.execute(title, command, installLogPath); err != nil {
		return err
	}
	return nil
}
