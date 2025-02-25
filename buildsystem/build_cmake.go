package buildsystem

import (
	"buildenv/pkg/cmd"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

func NewCMake(config BuildConfig, generator string) *cmake {
	// Set default generator if not specified.
	if generator == "" {
		switch runtime.GOOS {
		case "darwin":
			generator = "Xcode"
		case "linux":
			generator = "Unix Makefiles"
		case "windows":
			generator = "" // Let CMake choose the default Visual Studio generator.
		}
	}

	// Normalize generator name.
	switch strings.ToLower(generator) {
	case "ninja":
		generator = "Ninja"
	case "makefiles":
		generator = "Unix Makefiles"
	case "xcode":
		generator = "Xcode"
	default:
		generator = ""
	}

	return &cmake{
		BuildConfig: config,
		generator:   generator,
	}
}

type cmake struct {
	BuildConfig
	generator string // e.g. Ninja, Unix Makefiles, Visual Studio 16 2019, etc.
}

func (c cmake) Configure(buildType string) error {
	// Some libraries' configure or CMakeLists.txt may not in root folder.
	c.PortConfig.SourceDir = filepath.Join(c.PortConfig.SourceDir, c.PortConfig.SourceFolder)

	// Remove build dir and create it for configure.
	if err := os.RemoveAll(c.PortConfig.BuildDir); err != nil {
		return err
	}

	// Create build dir if not exists.
	if err := os.MkdirAll(c.PortConfig.BuildDir, os.ModeDir|os.ModePerm); err != nil {
		return err
	}

	// Override CMAKE_PREFIX_PATH and CMAKE_INSTALL_PREFIX.
	c.Options = slices.DeleteFunc(c.Options, func(element string) bool {
		return strings.Contains(element, "-DCMAKE_PREFIX_PATH=") ||
			strings.Contains(element, "-DCMAKE_INSTALL_PREFIX=")
	})
	c.Options = append(c.Options, fmt.Sprintf("-DCMAKE_PREFIX_PATH=%s", c.PortConfig.InstalledDir))
	c.Options = append(c.Options, fmt.Sprintf("-DCMAKE_INSTALL_PREFIX=%s", c.PortConfig.PackageDir))

	// Append cross-compile options only for none-runtime library.
	if !c.BuildConfig.AsDev {
		// Remove options that we want to override.
		c.Options = slices.DeleteFunc(c.Options, func(element string) bool {
			return strings.Contains(element, "-DCMAKE_POSITION_INDEPENDENT_CODE=") ||
				strings.Contains(element, "-DCMAKE_SYSTEM_PROCESSOR=") ||
				strings.Contains(element, "-DCMAKE_SYSTEM_NAME=") ||
				strings.Contains(element, "-DCMAKE_C_FLAGS=") ||
				strings.Contains(element, "-DCMAKE_CXX_FLAGS=") ||
				strings.Contains(element, "-DCMAKE_FIND_ROOT_PATH=") ||
				strings.Contains(element, "-DCMAKE_FIND_ROOT_PATH_MODE_PROGRAM=") ||
				strings.Contains(element, "-DCMAKE_FIND_ROOT_PATH_MODE_LIBRARY=") ||
				strings.Contains(element, "-DCMAKE_FIND_ROOT_PATH_MODE_INCLUDE=") ||
				strings.Contains(element, "-DCMAKE_FIND_ROOT_PATH_MODE_PACKAGE=")
		})

		// Append extra global args.
		c.Options = append(c.Options, fmt.Sprintf("-DCMAKE_POSITION_INDEPENDENT_CODE=%s", "ON"))

		c.Options = append(c.Options, fmt.Sprintf("-DCMAKE_SYSTEM_PROCESSOR=%s", c.PortConfig.CrossTools.SystemProcessor))
		c.Options = append(c.Options, fmt.Sprintf("-DCMAKE_SYSTEM_NAME=%s", c.PortConfig.CrossTools.SystemName))

		c.Options = append(c.Options, fmt.Sprintf("-DCMAKE_C_FLAGS=\"--sysroot=%s ${CMAKE_C_FLAGS}\"", c.PortConfig.CrossTools.RootFS))
		c.Options = append(c.Options, fmt.Sprintf("-DCMAKE_CXX_FLAGS=\"--sysroot=%s ${CMAKE_CXX_FLAGS}\"", c.PortConfig.CrossTools.RootFS))

		c.Options = append(c.Options, fmt.Sprintf("-DCMAKE_FIND_ROOT_PATH=\"%s\"", fmt.Sprintf("%s;%s",
			c.PortConfig.CrossTools.RootFS, c.PortConfig.InstalledDir)))
		c.Options = append(c.Options, fmt.Sprintf("-DCMAKE_FIND_ROOT_PATH_MODE_PROGRAM=%s", "NEVER"))
		c.Options = append(c.Options, fmt.Sprintf("-DCMAKE_FIND_ROOT_PATH_MODE_LIBRARY=%s", "ONLY"))
		c.Options = append(c.Options, fmt.Sprintf("-DCMAKE_FIND_ROOT_PATH_MODE_INCLUDE=%s", "ONLY"))
		c.Options = append(c.Options, fmt.Sprintf("-DCMAKE_FIND_ROOT_PATH_MODE_PACKAGE=%s", "ONLY"))
	}

	// Append 'CMAKE_BUILD_TYPE' if not contains it.
	if c.AsDev {
		c.Options = slices.DeleteFunc(c.Options, func(element string) bool {
			return strings.Contains(element, "CMAKE_BUILD_TYPE")
		})
		c.Options = append(c.Options, "-DCMAKE_BUILD_TYPE=Release")
	} else {
		if !slices.ContainsFunc(c.Options, func(arg string) bool {
			return strings.Contains(arg, "CMAKE_BUILD_TYPE")
		}) {
			buildType = c.formatBuildType(buildType)
			c.Options = append(c.Options, "-DCMAKE_BUILD_TYPE="+buildType)
		}
	}

	// Override library type if specified.
	if c.BuildConfig.LibraryType != "" {
		c.Options = slices.DeleteFunc(c.Options, func(element string) bool {
			return strings.Contains(element, "BUILD_SHARED_LIBS") ||
				strings.Contains(element, "BUILD_STATIC_LIBS")
		})

		switch c.BuildConfig.LibraryType {
		case "static":
			c.Options = append(c.Options, "-DBUILD_STATIC_LIBS=ON")
			c.Options = append(c.Options, "-DBUILD_SHARED_LIBS=OFF")

		case "shared":
			c.Options = append(c.Options, "-DBUILD_SHARED_LIBS=ON")
			c.Options = append(c.Options, "-DBUILD_STATIC_LIBS=OFF")
		}
	}

	// Assemble args into a single command string.
	joinedArgs := strings.Join(c.Options, " ")
	var command string
	if c.generator == "" {
		command = fmt.Sprintf("cmake -S %s -B %s %s", c.PortConfig.SourceDir, c.PortConfig.BuildDir, joinedArgs)
	} else {
		command = fmt.Sprintf("cmake -G %s -S %s -B %s %s", c.generator, c.PortConfig.SourceDir, c.PortConfig.BuildDir, joinedArgs)
	}

	// Execute configure.
	logPath := c.getLogPath("configure")
	title := fmt.Sprintf("[configure %s@%s]", c.PortConfig.LibName, c.PortConfig.LibVersion)
	executor := cmd.NewExecutor(title, command)
	executor.SetLogPath(logPath)
	if err := executor.Execute(); err != nil {
		return err
	}

	return nil
}

func (c cmake) Build() error {
	// Assemble command.
	command := fmt.Sprintf("cmake --build %s --parallel %d", c.PortConfig.BuildDir, c.PortConfig.JobNum)

	// Execute build.
	logPath := c.getLogPath("build")
	title := fmt.Sprintf("[build %s@%s]", c.PortConfig.LibName, c.PortConfig.LibVersion)
	executor := cmd.NewExecutor(title, command)
	executor.SetLogPath(logPath)
	if err := executor.Execute(); err != nil {
		return err
	}

	return nil
}

func (c cmake) Install() error {
	// Assemble command.
	command := fmt.Sprintf("cmake --install %s", c.PortConfig.BuildDir)

	// Execute install.
	logPath := c.getLogPath("install")
	title := fmt.Sprintf("[install %s@%s]", c.PortConfig.LibName, c.PortConfig.LibVersion)
	executor := cmd.NewExecutor(title, command)
	executor.SetLogPath(logPath)
	if err := executor.Execute(); err != nil {
		return err
	}

	return nil
}

func (cmake) formatBuildType(buildType string) string {
	switch strings.ToLower(buildType) {
	case "release":
		return "Release"

	case "debug":
		return "Debug"

	case "relwithdebinfo":
		return "RelWithDebInfo"

	case "minsizerel":
		return "MinSizeRel"

	default:
		return "Release"
	}
}
