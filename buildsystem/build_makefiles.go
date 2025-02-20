package buildsystem

import (
	"buildenv/pkg/cmd"
	"buildenv/pkg/fileio"
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

	// Some makefiles project may not support configure.
	configured bool
}

func (m *makefiles) Configure(buildType string) error {
	// Different Makefile projects set the build_type in inconsistent ways,
	// Fortunately, it can be configured through CFLAGS and CXXFLAGS.
	m.setBuildType(buildType)

	// Remove build dir and create it for configure process.
	if err := os.RemoveAll(m.PortConfig.BuildDir); err != nil {
		return err
	}

	// Create build dir if not exists.
	if err := os.MkdirAll(m.PortConfig.BuildDir, os.ModeDir|os.ModePerm); err != nil {
		return err
	}

	// Some libraries may not need to configure.
	if !fileio.PathExists(m.PortConfig.SourceDir+"/configure") &&
		!fileio.PathExists(m.PortConfig.SourceDir+"/Configure") &&
		!fileio.PathExists(m.PortConfig.SourceDir+"/autogen.sh") {
		return nil
	}

	// Execute autogen if exist.
	if fileio.PathExists(m.PortConfig.SourceDir + "/autogen.sh") {
		parentDir := filepath.Dir(m.PortConfig.BuildDir)
		fileName := filepath.Base(m.PortConfig.BuildDir) + "-autogen.log"
		logPath := filepath.Join(parentDir, fileName)
		title := fmt.Sprintf("[autogen %s]", m.PortConfig.LibName)
		executor := cmd.NewExecutor(title, "./autogen.sh")
		executor.SetLogPath(logPath)
		executor.SetWorkDir(m.PortConfig.SourceDir)
		if err := executor.Execute(); err != nil {
			return err
		}
	}

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

	// Fix --cache-file's path.
	cacheFileIndex := slices.IndexFunc(m.Arguments, func(element string) bool {
		return strings.HasPrefix(element, "--cache-file=")
	})
	if cacheFileIndex >= 0 {
		fileName := strings.Split(m.Arguments[cacheFileIndex], "=")[1]
		portDir := filepath.Join(m.PortConfig.PortsDir, m.PortConfig.LibName, fileName)
		m.Arguments[cacheFileIndex] = "--cache-file=" + portDir
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

	// Execute configure.
	configure := fmt.Sprintf("%s/%s %s", m.PortConfig.SourceDir, configureFile, joinedArgs)
	logPath := m.getLogPath("configure")
	title := fmt.Sprintf("[configure %s@%s]", m.PortConfig.LibName, m.PortConfig.LibVersion)
	executor := cmd.NewExecutor(title, configure)
	executor.SetLogPath(logPath)
	executor.SetWorkDir(m.PortConfig.BuildDir)
	if err := executor.Execute(); err != nil {
		return err
	}

	m.configured = true
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

	// Project that cannot configure always build in source.
	if m.configured {
		executor.SetWorkDir(m.PortConfig.BuildDir)
	} else {
		executor.SetWorkDir(m.PortConfig.SourceDir)
	}

	if err := executor.Execute(); err != nil {
		return err
	}

	return nil
}

func (m makefiles) Install() error {
	// Assemble command.
	var command string
	if m.configured {
		command = "make install"
	} else {
		m.Arguments = append(m.Arguments, "prefix="+m.PortConfig.PackageDir)
		joinedArgs := strings.Join(m.Arguments, " ")
		command = fmt.Sprintf("make install -C %s %s", m.PortConfig.SourceDir, joinedArgs)
	}

	// Execute install.
	logPath := m.getLogPath("install")
	title := fmt.Sprintf("[install %s@%s]", m.PortConfig.LibName, m.PortConfig.LibVersion)
	executor := cmd.NewExecutor(title, command)
	executor.SetLogPath(logPath)

	if m.configured {
		executor.SetWorkDir(m.PortConfig.BuildDir)
	} else {
		executor.SetWorkDir(m.PortConfig.SourceDir)
	}

	if err := executor.Execute(); err != nil {
		return err
	}

	return nil
}
