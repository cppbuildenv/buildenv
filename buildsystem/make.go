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

	// Replace placeholders with real paths.
	for index, argument := range m.Arguments {
		if strings.Contains(argument, "${INSTALLED_DIR}") {
			m.Arguments[index] = strings.ReplaceAll(argument, "${INSTALLED_DIR}", m.InstalledDir)
		}

		if strings.Contains(argument, "${TOOLCHAIN_PREFIX}") {
			toolchainPrefix := os.Getenv("TOOLCHAIN_PREFIX")
			m.Arguments[index] = strings.ReplaceAll(argument, "${TOOLCHAIN_PREFIX}", toolchainPrefix)
		}

		if strings.Contains(argument, "${SYSROOT}") {
			sysroot := os.Getenv("SYSROOT")
			m.Arguments[index] = strings.ReplaceAll(argument, "${SYSROOT}", sysroot)
		}
	}

	// Assemble script.
	m.Arguments = append(m.Arguments, fmt.Sprintf("--prefix=%s", m.InstalledDir))

	// Assemble args into a string.
	joinedArgs := strings.Join(m.Arguments, " ")
	configure := fmt.Sprintf("%s/configure %s", m.SourceDir, joinedArgs)

	// Print process log.
	fmt.Printf("\n[BuildEnv]: %s\n", configure)

	// Execute configure.
	configureLogPath := filepath.Join(filepath.Dir(m.BuildDir), filepath.Base(m.BuildDir)+"-configure.log")
	if err := m.execute(configure, configureLogPath); err != nil {
		return err
	}

	return nil
}

func (m make) Build() error {
	// Assemble script.
	command := fmt.Sprintf("make -j %d", m.JobNum)

	// Print process log.
	fmt.Printf("\n[BuildEnv]: %s\n", command)

	// Execute build.
	buildLogPath := filepath.Join(filepath.Dir(m.BuildDir), filepath.Base(m.BuildDir)+"-build.log")
	if err := m.execute(command, buildLogPath); err != nil {
		return err
	}

	return nil
}

func (m make) Install() error {
	// Assemble script.
	command := "make install"

	// Print process log.
	fmt.Printf("\n[BuildEnv]: %s\n", command)

	// Execute install.
	installLogPath := filepath.Join(filepath.Dir(m.BuildDir), filepath.Base(m.BuildDir)+"-install.log")
	if err := m.execute(command, installLogPath); err != nil {
		return err
	}
	return nil
}
