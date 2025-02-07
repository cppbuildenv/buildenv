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
	// Prepare envs for configure.
	m.setupDependencyPaths()

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
	title := fmt.Sprintf("[configure %s]", m.PortConfig.LibName)
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
	title := fmt.Sprintf("[build %s]", m.PortConfig.LibName)
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
	title := fmt.Sprintf("[install %s]", m.PortConfig.LibName)
	executor := cmd.NewExecutor(title, command)
	executor.SetLogPath(logPath)
	if err := executor.Execute(); err != nil {
		return err
	}

	return nil
}

func (m makefiles) setupDependencyPaths() {
	// Some third-party libraries cannot find headers and libs with sysroot,
	// so we need to set CFLAGS, CXXFLAGS, LDFLAGS to let these third-party libaries find them.
	installedDir := m.BuildConfig.PortConfig.InstalledDir
	cflags := os.Getenv("CFLAGS")
	cxxflags := os.Getenv("CXXFLAGS")
	ldflags := os.Getenv("LDFLAGS")

	if strings.TrimSpace(cflags) == "" {
		os.Setenv("CFLAGS", fmt.Sprintf("-I%s/include", installedDir))
	} else {
		os.Setenv("CFLAGS", fmt.Sprintf("-I%s/include", installedDir)+" "+cflags)
	}
	if strings.TrimSpace(cxxflags) == "" {
		os.Setenv("CXXFLAGS", fmt.Sprintf("-I%s/include", installedDir))
	} else {
		os.Setenv("CXXFLAGS", fmt.Sprintf("-I%s/include", installedDir)+" "+cxxflags)
	}
	if strings.TrimSpace(ldflags) == "" {
		os.Setenv("LDFLAGS", fmt.Sprintf("-L%s/lib", installedDir))
	} else {
		os.Setenv("LDFLAGS", fmt.Sprintf("-L%s/lib", installedDir)+" "+ldflags)
	}

	// Append $PKG_CONFIG_PATH with pkgconfig path that in installed dir.
	pkgConfigPath := os.Getenv("PKG_CONFIG_PATH")
	if strings.TrimSpace(pkgConfigPath) == "" {
		os.Setenv("PKG_CONFIG_PATH", installedDir+"/lib/pkgconfig")
	} else {
		os.Setenv("PKG_CONFIG_PATH", installedDir+"/lib/pkgconfig"+string(os.PathListSeparator)+pkgConfigPath)
	}

	// We assume that pkg-config's sysroot is installedDir and change all pc file's prefix as "/".
	os.Setenv("PKG_CONFIG_SYSROOT_DIR", installedDir)
}
