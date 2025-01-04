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

	rootPaths := strings.Join(
		[]string{c.PortConfig.RootFS, c.PortConfig.InstalledDir},
		string(os.PathListSeparator),
	)

	// Append extra global args.
	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_PREFIX_PATH=%s", c.PortConfig.InstalledDir))
	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_INSTALL_PREFIX=%s", c.PortConfig.PackageDir))
	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_POSITION_INDEPENDENT_CODE=%s", "ON"))

	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_SYSTEM_PROCESSOR=%s", c.PortConfig.SystemProcessor))
	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_SYSTEM_NAME=%s", c.PortConfig.SystemName))

	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_C_FLAGS_INIT=--sysroot=%s", c.PortConfig.RootFS))
	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_CXX_FLAGS_INIT=--sysroot=%s", c.PortConfig.RootFS))

	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_FIND_ROOT_PATH=%s", rootPaths))
	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_FIND_ROOT_PATH_MODE_PROGRAM=%s", "NEVER"))
	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_FIND_ROOT_PATH_MODE_LIBRARY=%s", "ONLY"))
	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_FIND_ROOT_PATH_MODE_INCLUDE=%s", "ONLY"))
	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_FIND_ROOT_PATH_MODE_PACKAGE=%s", "ONLY"))

	// Append 'CMAKE_BUILD_TYPE' if not contains it.
	containBuildType := slices.ContainsFunc(c.Arguments, func(arg string) bool {
		return strings.Contains(arg, "CMAKE_BUILD_TYPE")
	})
	if !containBuildType {
		buildType = c.formatBuildType(buildType)
		c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_BUILD_TYPE=%s", buildType))
	}

	// Assemble args into a single command string.
	joinedArgs := strings.Join(c.Arguments, " ")
	sourceDir := filepath.Join(c.PortConfig.SourceDir, c.PortConfig.SourceFolder)
	configure := fmt.Sprintf("cmake -S %s -B %s %s", sourceDir, c.PortConfig.BuildDir, joinedArgs)

	// Execute configure.
	parentDir := filepath.Dir(c.PortConfig.BuildDir)
	fileName := filepath.Base(c.PortConfig.BuildDir) + "-configure.log"
	configureLogPath := filepath.Join(parentDir, fileName)
	title := fmt.Sprintf("[configure %s]", c.PortConfig.LibName)
	if err := c.execute(title, configure, configureLogPath); err != nil {
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
	if err := c.execute(title, command, buildLogPath); err != nil {
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
	if err := c.execute(title, command, installLogPath); err != nil {
		return err
	}

	return nil
}

func (c cmake) formatBuildType(buildType string) string {
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
