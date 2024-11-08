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
	if err := os.RemoveAll(c.BuildDir); err != nil {
		return err
	}
	if err := os.MkdirAll(c.BuildDir, os.ModeDir|os.ModePerm); err != nil {
		return err
	}

	// Assemble script.
	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_PREFIX_PATH=%s", c.InstalledDir))
	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_INSTALL_PREFIX=%s", c.InstalledDir))

	// Append 'CMAKE_BUILD_TYPE' if not contains it.
	containBuildType := slices.ContainsFunc(c.Arguments, func(arg string) bool {
		return strings.Contains(arg, "CMAKE_BUILD_TYPE")
	})
	if !containBuildType {
		buildType = c.formatBuildType(buildType)
		c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_BUILD_TYPE=%s", buildType))
	}

	// Assemble args into a string.
	joinedArgs := strings.Join(c.Arguments, " ")
	command := fmt.Sprintf("cmake -S %s -B %s %s", c.SourceDir, c.BuildDir, joinedArgs)

	// Print process log.
	fmt.Printf("%s\n\n", command)

	// Execute configure.
	configureLogPath := filepath.Join(filepath.Dir(c.BuildDir), filepath.Base(c.BuildDir)+"-configure.log")
	if err := c.execute(command, configureLogPath); err != nil {
		return err
	}

	return nil
}

func (c cmake) Build() error {
	// Assemble script.
	command := fmt.Sprintf("cmake --build %s --parallel %d", c.BuildDir, c.JobNum)

	// Print process log.
	fmt.Printf("%s\n\n", command)

	// Execute build.
	buildLogPath := filepath.Join(filepath.Dir(c.BuildDir), filepath.Base(c.BuildDir)+"-build.log")
	if err := c.execute(command, buildLogPath); err != nil {
		return err
	}

	return nil
}

func (c cmake) Install() error {
	// Assemble script.
	command := fmt.Sprintf("cmake --install %s", c.BuildDir)

	// Print process log.
	fmt.Printf("%s\n\n", command)

	// Execute install.
	installLogPath := filepath.Join(filepath.Dir(c.BuildDir), filepath.Base(c.BuildDir)+"-install.log")
	if err := c.execute(command, installLogPath); err != nil {
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
