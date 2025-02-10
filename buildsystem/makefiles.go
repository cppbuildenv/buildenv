package buildsystem

import (
	"buildenv/pkg/cmd"
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
	// Different Makefile projects set the build_type in inconsistent ways,
	// Fortunately, it can be configured through CFLAGS and CXXFLAGS.
	m.setBuildType(buildType)

	// Clear cross build envs when build as dev.
	if m.BuildConfig.AsDev {
		m.PortConfig.CrossTools.ClearEnvs()
	} else {
		m.PortConfig.CrossTools.SetEnvs()
	}

	// Remove build dir and create it for configure process.
	if err := os.RemoveAll(m.PortConfig.BuildDir); err != nil {
		return err
	}

	// Create build dir if not exists.
	if err := os.MkdirAll(m.PortConfig.BuildDir, os.ModeDir|os.ModePerm); err != nil {
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

	// Remove common cross compile args for native build.
	if m.PortConfig.CrossTools.Native || m.BuildConfig.AsDev {
		m.Arguments = slices.DeleteFunc(m.Arguments, func(element string) bool {
			return strings.Contains(element, "--host=") ||
				strings.Contains(element, "--sysroot=") ||
				strings.Contains(element, "--cross-prefix=") ||
				strings.Contains(element, "--enable-cross-compile") ||
				strings.Contains(element, "--arch=") ||
				strings.Contains(element, "--target-os=")
		})
	}

	joinedArgs := strings.Join(m.Arguments, " ")

	// Find `configure` or `Configure`.
	var configureFile string
	if _, err := os.Stat(m.PortConfig.SourceDir + "/configure"); err == nil {
		configureFile = "configure"
	}
	if _, err := os.Stat(m.PortConfig.SourceDir + "/Configure"); err == nil {
		configureFile = "Configure"
	}

	// Execute autogen.
	if configureFile == "" {
		if _, err := os.Stat(m.PortConfig.SourceDir + "/autogen.sh"); err == nil {
			if err := os.Chdir(m.PortConfig.SourceDir); err != nil {
				return err
			}

			parentDir := filepath.Dir(m.PortConfig.BuildDir)
			fileName := filepath.Base(m.PortConfig.BuildDir) + "-autogen.log"
			logPath := filepath.Join(parentDir, fileName)
			title := fmt.Sprintf("[autogen %s]", m.PortConfig.LibName)
			executor := cmd.NewExecutor(title, "./autogen.sh")
			executor.SetLogPath(logPath)
			if err := executor.Execute(); err != nil {
				return err
			}
		}
	}

	// Make sure create build cache in build dir.
	if err := os.Chdir(m.PortConfig.BuildDir); err != nil {
		return err
	}

	// Find `configure` or `Configure`.
	if _, err := os.Stat(m.PortConfig.SourceDir + "/configure"); err == nil {
		configureFile = "configure"
	}
	if _, err := os.Stat(m.PortConfig.SourceDir + "/Configure"); err == nil {
		configureFile = "Configure"
	}

	// Execute configure.
	configure := fmt.Sprintf("%s/%s %s", m.PortConfig.SourceDir, configureFile, joinedArgs)
	logPath := m.getLogPath("configure")
	title := fmt.Sprintf("[configure %s@%s]", m.PortConfig.LibName, m.PortConfig.LibVersion)
	executor := cmd.NewExecutor(title, configure)
	executor.SetLogPath(logPath)
	if err := executor.Execute(); err != nil {
		return err
	}

	return nil
}

func (m makefiles) Build() error {
	// Assemble command.
	command := fmt.Sprintf("make -j %d", m.PortConfig.JobNum)

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

func (m makefiles) Install() error {
	// Assemble command.
	command := "make install"

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
