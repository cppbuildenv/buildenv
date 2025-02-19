package buildsystem

import (
	"buildenv/pkg/cmd"
	"bytes"
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

	// Create build dir if not exists.
	if err := os.MkdirAll(m.PortConfig.BuildDir, os.ModeDir|os.ModePerm); err != nil {
		return err
	}

	// Override '--prefix' if exists.
	m.Arguments = slices.DeleteFunc(m.Arguments, func(element string) bool {
		return strings.Contains(element, "--prefix")
	})
	m.Arguments = append(m.Arguments, "--prefix="+m.PortConfig.PackageDir)

	// Append 'CMAKE_BUILD_TYPE' if not contains it.
	if m.AsDev {
		m.Arguments = slices.DeleteFunc(m.Arguments, func(element string) bool {
			return strings.Contains(element, "--buildtype")
		})
		m.Arguments = append(m.Arguments, "--buildtype=release")
	} else {
		if !slices.ContainsFunc(m.Arguments, func(arg string) bool {
			return strings.Contains(arg, "--buildtype")
		}) {
			buildType = strings.ToLower(buildType)
			m.Arguments = append(m.Arguments, "--buildtype="+buildType)
		}
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

	// Assemble command.
	var command string
	joinedArgs := strings.Join(m.Arguments, " ")
	if m.BuildConfig.AsDev {
		command = fmt.Sprintf("meson setup %s %s", m.PortConfig.BuildDir, joinedArgs)
	} else {
		crossFile, err := m.generateCrossFile()
		if err != nil {
			return fmt.Errorf("failed to generate cross_file.txt for meson: %v", err)
		}
		command = fmt.Sprintf("meson setup %s %s --cross-file %s", m.PortConfig.BuildDir, joinedArgs, crossFile)
	}

	// Execute configure.
	logPath := m.getLogPath("configure")
	title := fmt.Sprintf("[configure %s@%s]", m.PortConfig.LibName, m.PortConfig.LibVersion)
	executor := cmd.NewExecutor(title, command)
	executor.SetLogPath(logPath)
	if err := executor.Execute(); err != nil {
		return err
	}

	return nil
}

func (m meson) Build() error {
	// Assemble command.
	command := fmt.Sprintf("meson compile -C %s -j %d", m.PortConfig.BuildDir, m.PortConfig.JobNum)

	// Execute build.
	logPath := m.getLogPath("build")
	title := fmt.Sprintf("[build %s@%s]", m.PortConfig.LibName, m.PortConfig.LibVersion)
	executor := cmd.NewExecutor(title, command)
	executor.SetLogPath(logPath)
	if err := executor.Execute(); err != nil {
		return err
	}

	return nil
}

func (m meson) Install() error {
	// Assemble command.
	command := fmt.Sprintf("meson install -C %s", m.PortConfig.BuildDir)

	// Execute install.
	logPath := m.getLogPath("install")
	title := fmt.Sprintf("[install %s@%s]", m.PortConfig.LibName, m.PortConfig.LibVersion)
	executor := cmd.NewExecutor(title, command)
	executor.SetLogPath(logPath)
	if err := executor.Execute(); err != nil {
		return err
	}

	return nil
}

func (m meson) generateCrossFile() (string, error) {
	var bytes bytes.Buffer
	bytes.WriteString("[host_machine]\n")
	bytes.WriteString(fmt.Sprintf("system = '%s'\n", m.PortConfig.CrossTools.SystemName))
	bytes.WriteString(fmt.Sprintf("cpu_family = '%s'\n", m.PortConfig.CrossTools.SystemProcessor))
	bytes.WriteString(fmt.Sprintf("cpu = '%s'\n", m.PortConfig.CrossTools.SystemProcessor))
	bytes.WriteString("endian = 'little'\n")

	bytes.WriteString("\n[binaries]\n")
	bytes.WriteString(fmt.Sprintf("c = '%s'\n", m.PortConfig.CrossTools.CC))
	bytes.WriteString(fmt.Sprintf("cpp = '%s'\n", m.PortConfig.CrossTools.CXX))

	if m.PortConfig.CrossTools.FC != "" {
		bytes.WriteString(fmt.Sprintf("fc = '%s'\n", m.PortConfig.CrossTools.FC))
	}
	if m.PortConfig.CrossTools.RANLIB != "" {
		bytes.WriteString(fmt.Sprintf("ranlib = '%s'\n", m.PortConfig.CrossTools.RANLIB))
	}
	if m.PortConfig.CrossTools.AR != "" {
		bytes.WriteString(fmt.Sprintf("ar = '%s'\n", m.PortConfig.CrossTools.AR))
	}
	if m.PortConfig.CrossTools.LD != "" {
		bytes.WriteString(fmt.Sprintf("ld = '%s'\n", m.PortConfig.CrossTools.LD))
	}
	if m.PortConfig.CrossTools.NM != "" {
		bytes.WriteString(fmt.Sprintf("nm = '%s'\n", m.PortConfig.CrossTools.NM))
	}
	if m.PortConfig.CrossTools.OBJDUMP != "" {
		bytes.WriteString(fmt.Sprintf("objdump = '%s'\n", m.PortConfig.CrossTools.OBJDUMP))
	}
	if m.PortConfig.CrossTools.STRIP != "" {
		bytes.WriteString(fmt.Sprintf("strip = '%s'\n", m.PortConfig.CrossTools.STRIP))
	}

	bytes.WriteString("\n[properties]\n")
	bytes.WriteString("cross_file = 'true'\n")
	bytes.WriteString(fmt.Sprintf("sys_root = '%s'\n", m.PortConfig.CrossTools.RootFS))
	bytes.WriteString(fmt.Sprintf("pkg_config_libdir = '%s'\n", os.Getenv("PKG_CONFIG_LIBDIR")))

	crossFilePath := filepath.Join(m.PortConfig.BuildDir, "cross_file.txt")
	if err := os.WriteFile(crossFilePath, bytes.Bytes(), os.ModePerm); err != nil {
		return "", err
	}

	return crossFilePath, nil
}
