package buildsystem

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func NewMakefiles(config BuildConfig) *makefiles {
	return &makefiles{BuildConfig: config}
}

type makefiles struct {
	BuildConfig
}

func (m makefiles) Configure(buildType string) error {
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

	// Some third-party's configure scripts is not exist in the source folder root.
	m.PortConfig.SourceDir = filepath.Join(m.PortConfig.SourceDir, m.PortConfig.SourceFolder)

	// Append common variables for cross compiling.
	m.Arguments = append(m.Arguments, fmt.Sprintf("--prefix=%s", m.PortConfig.PackageDir))

	// Override library type if specified.
	if m.BuildConfig.LibraryType != "" {
		m.Arguments = slices.DeleteFunc(m.Arguments, func(element string) bool {
			return strings.Contains(element, "--enable-shared") ||
				strings.Contains(element, "--disable-shared") ||
				strings.Contains(element, "--enable-static") ||
				strings.Contains(element, "--disable-static")
		})

		switch m.BuildConfig.LibraryType {
		case "static":
			m.Arguments = append(m.Arguments, "--enable-static")
			m.Arguments = append(m.Arguments, "--disable-shared")

		case "shared":
			m.Arguments = append(m.Arguments, "--enable-shared")
			m.Arguments = append(m.Arguments, "--disable-static")
		}
	}

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

	// Execute autogen.
	if m.BuildConfig.AutogenConfigure {
		autogen := fmt.Sprintf("%s/autogen.sh", m.PortConfig.SourceDir)
		parentDir := filepath.Dir(m.PortConfig.BuildDir)
		fileName := filepath.Base(m.PortConfig.BuildDir) + "-autogen.log"
		configureLogPath := filepath.Join(parentDir, fileName)
		title := fmt.Sprintf("[configure %s]", m.PortConfig.LibName)
		if err := execute(title, autogen, configureLogPath); err != nil {
			return err
		}
	}

	// Find `configure` or `Configure`.
	var configureFile string
	if _, err := os.Stat(m.PortConfig.SourceDir + "/configure"); err == nil {
		configureFile = "configure"
	}
	if _, err := os.Stat(m.PortConfig.SourceDir + "/Configure"); err == nil {
		configureFile = "Configure"
	}

	// Join args into a string.
	joinedArgs := strings.Join(m.Arguments, " ")
	configure := fmt.Sprintf("%s/%s %s", m.PortConfig.SourceDir, configureFile, joinedArgs)

	// Execute configure.
	parentDir := filepath.Dir(m.PortConfig.BuildDir)
	fileName := filepath.Base(m.PortConfig.BuildDir) + "-configure.log"
	configureLogPath := filepath.Join(parentDir, fileName)
	title := fmt.Sprintf("[configure %s]", m.PortConfig.LibName)
	if err := execute(title, configure, configureLogPath); err != nil {
		return err
	}

	return nil
}

func (m makefiles) Build() error {
	// Assemble script.
	command := fmt.Sprintf("make -j %d", m.PortConfig.JobNum)

	// Execute build.
	parentDir := filepath.Dir(m.PortConfig.BuildDir)
	fileName := filepath.Base(m.PortConfig.BuildDir) + "-build.log"
	buildLogPath := filepath.Join(parentDir, fileName)
	title := fmt.Sprintf("[build %s]", m.PortConfig.LibName)
	if err := execute(title, command, buildLogPath); err != nil {
		return err
	}

	return nil
}

func (m makefiles) Install() error {
	// Assemble script.
	command := "make install"

	// Execute install.
	parentDir := filepath.Dir(m.PortConfig.BuildDir)
	fileName := filepath.Base(m.PortConfig.BuildDir) + "-install.log"
	installLogPath := filepath.Join(parentDir, fileName)
	title := fmt.Sprintf("[install %s]", m.PortConfig.LibName)
	if err := execute(title, command, installLogPath); err != nil {
		return err
	}

	return nil
}
