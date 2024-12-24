package buildsystem

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

func NewCMake(config BuildConfig) *cmake {
	return &cmake{BuildConfig: config}
}

type cmake struct {
	BuildConfig
}

func (c cmake) Configure(buildType string) (string, error) {
	// Remove build dir and create it for configure.
	if err := os.RemoveAll(c.portConfig.BuildDir); err != nil {
		return "", err
	}
	if err := os.MkdirAll(c.portConfig.BuildDir, os.ModeDir|os.ModePerm); err != nil {
		return "", err
	}

	// Append extra global args.
	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_PREFIX_PATH=%s", c.portConfig.InstalledDir))
	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_INSTALL_PREFIX=%s", c.portConfig.InstalledDir))
	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_POSITION_INDEPENDENT_CODE=%s", "ON"))

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
	sourceDir := filepath.Join(c.portConfig.SourceDir, c.portConfig.SourceFolder)
	configure := fmt.Sprintf("cmake -S %s -B %s %s", sourceDir, c.portConfig.BuildDir, joinedArgs)

	// Execute configure.
	parentDir := filepath.Dir(c.portConfig.BuildDir)
	fileName := filepath.Base(c.portConfig.BuildDir) + "-configure.log"
	configureLogPath := filepath.Join(parentDir, fileName)
	title := fmt.Sprintf("[configure %s]", c.portConfig.LibName)
	if err := c.execute(title, configure, configureLogPath); err != nil {
		return "", err
	}

	return configureLogPath, nil
}

func (c cmake) Build() (string, error) {
	// Assemble script.
	command := fmt.Sprintf("cmake --build %s --parallel %d", c.portConfig.BuildDir, c.portConfig.JobNum)

	// Execute build.
	parentDir := filepath.Dir(c.portConfig.BuildDir)
	fileName := filepath.Base(c.portConfig.BuildDir) + "-build.log"
	buildLogPath := filepath.Join(parentDir, fileName)
	title := fmt.Sprintf("[build %s]", c.portConfig.LibName)
	if err := c.execute(title, command, buildLogPath); err != nil {
		return "", err
	}

	return buildLogPath, nil
}

func (c cmake) Install() (string, error) {
	// Assemble script.
	command := fmt.Sprintf("cmake --install %s", c.portConfig.BuildDir)

	// Execute install.
	parentDir := filepath.Dir(c.portConfig.BuildDir)
	fileName := filepath.Base(c.portConfig.BuildDir) + "-install.log"
	installLogPath := filepath.Join(parentDir, fileName)
	title := fmt.Sprintf("[install %s]", c.portConfig.LibName)
	if err := c.execute(title, command, installLogPath); err != nil {
		return "", err
	}

	return installLogPath, nil
}

func (c cmake) InstalledFiles(installLogFile string) ([]string, error) {
	file, err := os.OpenFile(installLogFile, os.O_RDONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var files []string                                               // All installed files.
	reg := regexp.MustCompile(`^-- (Installing:|Up-to-date:) (\S+)`) // Installed file regex.

	// Read line by line to find installed files.
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		match := reg.FindStringSubmatch(line)

		if len(match) > 2 {
			installedFile := match[2]
			installedFile = strings.TrimPrefix(installedFile, c.portConfig.InstalledRootDir+"/")
			files = append(files, installedFile)
		}
	}

	return files, nil
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
