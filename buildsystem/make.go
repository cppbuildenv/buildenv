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

func NewMake(config BuildConfig) *make {
	return &make{BuildConfig: config}
}

type make struct {
	BuildConfig
}

func (m make) Configure(buildType string) (string, error) {
	// Remove build dir and create it for configure.
	if err := os.RemoveAll(m.portConfig.BuildDir); err != nil {
		return "", err
	}
	if err := os.MkdirAll(m.portConfig.BuildDir, os.ModeDir|os.ModePerm); err != nil {
		return "", err
	}

	if err := os.Chdir(m.portConfig.BuildDir); err != nil {
		return "", err
	}

	// Append common variables for cross compiling.
	m.Arguments = append(m.Arguments, fmt.Sprintf("--prefix=%s", m.portConfig.InstalledDir))

	// Replace placeholders with real paths.
	for index, argument := range m.Arguments {
		if strings.Contains(argument, "${HOST}") {
			m.Arguments[index] = strings.ReplaceAll(argument, "${HOST}", m.portConfig.Host)
		}

		if strings.Contains(argument, "${SYSTEM_NAME}") {
			m.Arguments[index] = strings.ReplaceAll(argument, "${SYSTEM_NAME}", strings.ToLower(m.portConfig.SystemName))
		}

		if strings.Contains(argument, "${SYSTEM_PROCESSOR}") {
			m.Arguments[index] = strings.ReplaceAll(argument, "${SYSTEM_PROCESSOR}", m.portConfig.SystemProcessor)
		}

		if strings.Contains(argument, "${SYSROOT}") {
			m.Arguments[index] = strings.ReplaceAll(argument, "${SYSROOT}", m.portConfig.RootFS)
		}

		if strings.Contains(argument, "${CROSS_PREFIX}") {
			m.Arguments[index] = strings.ReplaceAll(argument, "${CROSS_PREFIX}", m.portConfig.ToolchainPrefix)
		}
	}

	// Join args into a string.
	joinedArgs := strings.Join(m.Arguments, " ")
	configure := fmt.Sprintf("%s/configure %s", m.portConfig.SourceDir, joinedArgs)

	// Execute configure.
	parentDir := filepath.Dir(m.portConfig.BuildDir)
	fileName := filepath.Base(m.portConfig.BuildDir) + "-configure.log"
	configureLogPath := filepath.Join(parentDir, fileName)
	title := fmt.Sprintf("[configure %s]", m.portConfig.LibName)
	if err := m.execute(title, configure, configureLogPath); err != nil {
		return "", err
	}

	return configureLogPath, nil
}

func (m make) Build() (string, error) {
	// Assemble script.
	command := fmt.Sprintf("make -j %d", m.portConfig.JobNum)

	// Execute build.
	parentDir := filepath.Dir(m.portConfig.BuildDir)
	fileName := filepath.Base(m.portConfig.BuildDir) + "-build.log"
	buildLogPath := filepath.Join(parentDir, fileName)
	title := fmt.Sprintf("[build %s]", m.portConfig.LibName)
	if err := m.execute(title, command, buildLogPath); err != nil {
		return "", err
	}

	return buildLogPath, nil
}

func (m make) Install() (string, error) {
	// Assemble script.
	command := "make install"

	// Execute install.
	parentDir := filepath.Dir(m.portConfig.BuildDir)
	fileName := filepath.Base(m.portConfig.BuildDir) + "-install.log"
	installLogPath := filepath.Join(parentDir, fileName)
	title := fmt.Sprintf("[install %s]", m.portConfig.LibName)
	if err := m.execute(title, command, installLogPath); err != nil {
		return "", err
	}
	return installLogPath, nil
}

func (m make) InstalledFiles(installLogFile string) ([]string, error) {
	file, err := os.OpenFile(installLogFile, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	installReg1 := regexp.MustCompile(`^install -m \d{3} (.+)`) // case like x264
	installReg2 := regexp.MustCompile(`^INSTALL\s+(.+)`)        // case like ffmpeg
	lnReg := regexp.MustCompile(`^ln\s+-f\s+-s? (.+) (.+)$`)    // case like x264

	var installedFiles []string

	// Read line by line to find installed files.
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		matchInstall1 := installReg1.FindStringSubmatch(line)
		matchInstall2 := installReg2.FindStringSubmatch(line)
		matchLn := lnReg.FindStringSubmatch(line)

		if len(matchInstall1) > 1 || len(matchInstall2) > 1 {
			var line string

			if len(matchInstall1) > 1 {
				line = matchInstall1[1]
			} else {
				line = matchInstall2[1]
			}

			segments := strings.Split(line, " ")
			for _, segment := range segments {
				if strings.HasPrefix(segment, m.portConfig.InstalledDir) {
					continue
				}

				parentFolder := filepath.Base(filepath.Dir(segment))
				path, err := m.findInstalledFile(parentFolder, filepath.Base(segment))
				if err != nil {
					fmt.Printf("can not find installed file: %s\n", segment)
					continue
				}

				// Some makefile install may contains duplicated files.
				path = strings.TrimPrefix(path, m.portConfig.InstalledRootDir+"/")
				if slices.Index(installedFiles, path) == -1 {
					installedFiles = append(installedFiles, path)
				}
			}
		} else if len(matchLn) > 2 {
			line := matchLn[2]
			parentFolder := filepath.Base(filepath.Dir(line))
			path, err := m.findInstalledFile(parentFolder, filepath.Base(line))
			if err != nil {
				return nil, err
			}

			path = strings.TrimPrefix(path, m.portConfig.InstalledRootDir+"/")
			installedFiles = append(installedFiles, path)
		}
	}

	return installedFiles, nil
}

func (m make) findInstalledFile(parentDir, filename string) (string, error) {
	var filePaths []string

	if err := filepath.Walk(m.portConfig.InstalledDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == filename {
			filePaths = append(filePaths, path)
		}

		return nil
	}); err != nil {
		return "", err
	}

	if len(filePaths) == 1 {
		return filePaths[0], nil
	}

	// Make sure both file and parent dir are matched.
	for _, path := range filePaths {
		if strings.HasSuffix(path, filepath.Join(parentDir, filename)) {
			return path, nil
		}
	}

	return "", fmt.Errorf("file %s not found", filename)
}
