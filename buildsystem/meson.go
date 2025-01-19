package buildsystem

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func NewMeson(config BuildConfig) *meson {
	return &meson{BuildConfig: config}
}

type meson struct {
	BuildConfig
}

func (m meson) Configure(buildType string) error {
	// Remove build dir and create it for configure.
	if err := os.RemoveAll(m.PortConfig.BuildDir); err != nil {
		return err
	}
	if err := os.MkdirAll(m.PortConfig.BuildDir, os.ModeDir|os.ModePerm); err != nil {
		return err
	}

	// Some third-party's configure scripts is not exist in the source folder root.
	m.PortConfig.SourceDir = filepath.Join(m.PortConfig.SourceDir, m.PortConfig.SourceFolder)

	// Override '--prefix' if exists.
	m.Arguments = slices.DeleteFunc(m.Arguments, func(element string) bool {
		return strings.Contains(element, "--prefix")
	})
	m.Arguments = append(m.Arguments, "--prefix="+m.PortConfig.PackageDir)

	// Append 'CMAKE_BUILD_TYPE' if not contains it.
	if !slices.ContainsFunc(m.Arguments, func(arg string) bool {
		return strings.Contains(arg, "--buildtype")
	}) {
		buildType = m.FormatBuildType(buildType)
		m.Arguments = append(m.Arguments, "--buildtype="+buildType)
	}

	// Override library type if specified.
	if m.BuildConfig.LibraryType != "" {
		m.Arguments = slices.DeleteFunc(m.Arguments, func(element string) bool {
			return strings.Contains(element, "--default-library")
		})

		switch m.BuildConfig.LibraryType {
		case "static":
			m.Arguments = append(m.Arguments, "--default-library=static")

		case "shared":
			m.Arguments = append(m.Arguments, "--default-library=shared")
		}
	}

	if err := os.Chdir(m.PortConfig.SourceDir); err != nil {
		return err
	}

	// Assemble args into a single command string.
	joinedArgs := strings.Join(m.Arguments, " ")
	configure := fmt.Sprintf("meson setup %s %s", m.PortConfig.BuildDir, joinedArgs)

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

func (m meson) Build() error {
	// Assemble script.
	command := fmt.Sprintf("ninja -C %s -j %d", m.PortConfig.BuildDir, m.PortConfig.JobNum)

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

func (m meson) Install() error {
	// Assemble script.
	command := fmt.Sprintf("ninja -C %s install", m.PortConfig.BuildDir)

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

func (m meson) FormatBuildType(buildType string) string {
	return strings.ToLower(buildType)
}
