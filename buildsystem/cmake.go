package buildsystem

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func NewCMake(config BuildConfig) *cmake {
	return &cmake{BuildConfig: config}
}

type cmake struct {
	BuildConfig
}

func (c cmake) Configure(buildType string) error {
	// Remove build dir and create it for configure.
	if err := os.RemoveAll(c.PortConfig.BuildDir); err != nil {
		return err
	}
	if err := os.MkdirAll(c.PortConfig.BuildDir, os.ModeDir|os.ModePerm); err != nil {
		return err
	}

	// Some third-party's configure scripts is not exist in the source folder root.
	c.PortConfig.SourceDir = filepath.Join(c.PortConfig.SourceDir, c.PortConfig.SourceFolder)

	// Remove arguments that we want to override.
	c.Arguments = slices.DeleteFunc(c.Arguments, func(element string) bool {
		return strings.Contains(element, "-DCMAKE_PREFIX_PATH=") ||
			strings.Contains(element, "-DCMAKE_INSTALL_PREFIX=") ||
			strings.Contains(element, "-DCMAKE_POSITION_INDEPENDENT_CODE=") ||
			strings.Contains(element, "-DCMAKE_SYSTEM_PROCESSOR=") ||
			strings.Contains(element, "-DCMAKE_SYSTEM_NAME=") ||
			strings.Contains(element, "-DCMAKE_C_FLAGS_INIT=") ||
			strings.Contains(element, "-DCMAKE_CXX_FLAGS_INIT=") ||
			strings.Contains(element, "-DCMAKE_FIND_ROOT_PATH=") ||
			strings.Contains(element, "-DCMAKE_FIND_ROOT_PATH_MODE_PROGRAM=") ||
			strings.Contains(element, "-DCMAKE_FIND_ROOT_PATH_MODE_LIBRARY=") ||
			strings.Contains(element, "-DCMAKE_FIND_ROOT_PATH_MODE_INCLUDE=") ||
			strings.Contains(element, "-DCMAKE_FIND_ROOT_PATH_MODE_PACKAGE=")
	})

	// Append extra global args.
	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_PREFIX_PATH=%s", c.PortConfig.InstalledDir))
	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_INSTALL_PREFIX=%s", c.PortConfig.PackageDir))
	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_POSITION_INDEPENDENT_CODE=%s", "ON"))

	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_SYSTEM_PROCESSOR=%s", c.PortConfig.SystemProcessor))
	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_SYSTEM_NAME=%s", c.PortConfig.SystemName))

	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_C_FLAGS_INIT=--sysroot=%s", c.PortConfig.RootFS))
	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_CXX_FLAGS_INIT=--sysroot=%s", c.PortConfig.RootFS))

	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_FIND_ROOT_PATH=%s",
		strings.Join([]string{c.PortConfig.RootFS, c.PortConfig.InstalledDir}, string(os.PathListSeparator))))
	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_FIND_ROOT_PATH_MODE_PROGRAM=%s", "NEVER"))
	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_FIND_ROOT_PATH_MODE_LIBRARY=%s", "ONLY"))
	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_FIND_ROOT_PATH_MODE_INCLUDE=%s", "ONLY"))
	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_FIND_ROOT_PATH_MODE_PACKAGE=%s", "ONLY"))

	// Append 'CMAKE_BUILD_TYPE' if not contains it.
	if !slices.ContainsFunc(c.Arguments, func(arg string) bool {
		return strings.Contains(arg, "CMAKE_BUILD_TYPE")
	}) {
		buildType = c.FormatBuildType(buildType)
		c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_BUILD_TYPE=%s", buildType))
	}

	// Override library type if specified.
	if c.BuildConfig.LibraryType != "" {
		c.Arguments = slices.DeleteFunc(c.Arguments, func(element string) bool {
			return strings.Contains(element, "BUILD_SHARED_LIBS") ||
				strings.Contains(element, "BUILD_STATIC_LIBS")
		})

		switch c.BuildConfig.LibraryType {
		case "static":
			c.Arguments = append(c.Arguments, "-DBUILD_STATIC_LIBS=ON")
			c.Arguments = append(c.Arguments, "-DBUILD_SHARED_LIBS=OFF")

		case "shared":
			c.Arguments = append(c.Arguments, "-DBUILD_SHARED_LIBS=ON")
			c.Arguments = append(c.Arguments, "-DBUILD_STATIC_LIBS=OFF")
		}
	}

	// Assemble args into a single command string.
	joinedArgs := strings.Join(c.Arguments, " ")
	configure := fmt.Sprintf("cmake -S %s -B %s %s", c.PortConfig.SourceDir, c.PortConfig.BuildDir, joinedArgs)

	// Execute configure.
	parentDir := filepath.Dir(c.PortConfig.BuildDir)
	fileName := filepath.Base(c.PortConfig.BuildDir) + "-configure.log"
	configureLogPath := filepath.Join(parentDir, fileName)
	title := fmt.Sprintf("[configure %s]", c.PortConfig.LibName)
	if err := execute(title, configure, configureLogPath); err != nil {
		return err
	}

	return nil
}

func (c cmake) Build() error {
	// Assemble script.
	command := fmt.Sprintf("cmake --build %s --parallel %d", c.PortConfig.BuildDir, c.PortConfig.JobNum)

	// Execute build.
	parentDir := filepath.Dir(c.PortConfig.BuildDir)
	fileName := filepath.Base(c.PortConfig.BuildDir) + "-build.log"
	buildLogPath := filepath.Join(parentDir, fileName)
	title := fmt.Sprintf("[build %s]", c.PortConfig.LibName)
	if err := execute(title, command, buildLogPath); err != nil {
		return err
	}

	return nil
}

func (c cmake) Install() error {
	// Assemble script.
	command := fmt.Sprintf("cmake --install %s", c.PortConfig.BuildDir)

	// Execute install.
	parentDir := filepath.Dir(c.PortConfig.BuildDir)
	fileName := filepath.Base(c.PortConfig.BuildDir) + "-install.log"
	installLogPath := filepath.Join(parentDir, fileName)
	title := fmt.Sprintf("[install %s]", c.PortConfig.LibName)
	if err := execute(title, command, installLogPath); err != nil {
		return err
	}

	return nil
}

func (c cmake) FormatBuildType(buildType string) string {
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
