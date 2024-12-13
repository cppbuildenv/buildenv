package buildsystem

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	if err := os.RemoveAll(m.BuildDir); err != nil {
		return "", err
	}
	if err := os.MkdirAll(m.BuildDir, os.ModeDir|os.ModePerm); err != nil {
		return "", err
	}

	if err := os.Chdir(m.BuildDir); err != nil {
		return "", err
	}

	var (
		crossPrefix = os.Getenv("CROSS_PREFIX")
		sysroot     = os.Getenv("SYSROOT")
		host        = os.Getenv("HOST")
	)

	// Append common variables for cross compiling.
	m.Arguments = append(m.Arguments, fmt.Sprintf("--prefix=%s", m.InstalledDir))
	m.Arguments = append(m.Arguments, fmt.Sprintf("--sysroot=%s", sysroot))
	m.Arguments = append(m.Arguments, fmt.Sprintf("--cross-prefix=%s", crossPrefix))

	// Replace placeholders with real paths.
	for index, argument := range m.Arguments {
		if strings.Contains(argument, "${INSTALLED_DIR}") {
			m.Arguments[index] = strings.ReplaceAll(argument, "${INSTALLED_DIR}", m.InstalledDir)
		}

		if strings.Contains(argument, "${HOST}") {
			m.Arguments[index] = strings.ReplaceAll(argument, "${HOST}", host)
		}
	}

	// Join args into a string.
	joinedArgs := strings.Join(m.Arguments, " ")
	configure := fmt.Sprintf("%s/configure %s", m.SourceDir, joinedArgs)

	// Execute configure.
	configureLogPath := filepath.Join(filepath.Dir(m.BuildDir), filepath.Base(m.BuildDir)+"-configure.log")
	title := fmt.Sprintf("[configure %s]", m.LibName)
	if err := m.execute(title, configure, configureLogPath); err != nil {
		return "", err
	}

	return configureLogPath, nil
}

func (m make) Build() (string, error) {
	// Assemble script.
	command := fmt.Sprintf("make -j %d", m.JobNum)

	// Execute build.
	buildLogPath := filepath.Join(filepath.Dir(m.BuildDir), filepath.Base(m.BuildDir)+"-build.log")
	title := fmt.Sprintf("[build %s]", m.LibName)
	if err := m.execute(title, command, buildLogPath); err != nil {
		return "", err
	}

	return buildLogPath, nil
}

func (m make) Install() (string, error) {
	// Assemble script.
	command := "make install"

	// Execute install.
	installLogPath := filepath.Join(filepath.Dir(m.BuildDir), filepath.Base(m.BuildDir)+"-install.log")
	title := fmt.Sprintf("[install %s]", m.LibName)
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

	var files []string // All installed files.
	installReg1 := regexp.MustCompile(`^install -m \d{3} (.+)`)
	installReg2 := regexp.MustCompile(`^INSTALL\s+(.+)`)
	lnReg := regexp.MustCompile(`^ln\s+-f\s+-s? (.+) (.+)$`)

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
				if strings.HasPrefix(segment, m.InstalledDir) {
					continue
				}

				parentFolder := filepath.Base(filepath.Dir(segment))
				path, err := m.findInstalledFile(parentFolder, filepath.Base(segment))
				if err != nil {
					fmt.Printf("can not find installed file: %s\n", segment)
					continue
				}

				path = strings.TrimPrefix(path, m.InstalledRootDir+"/")
				files = append(files, path)
			}
		} else if len(matchLn) > 2 {
			line := matchLn[2]
			parentFolder := filepath.Base(filepath.Dir(line))
			path, err := m.findInstalledFile(parentFolder, filepath.Base(line))
			if err != nil {
				return nil, err
			}

			path = strings.TrimPrefix(path, m.InstalledRootDir+"/")
			files = append(files, path)
		}
	}

	return files, nil
}

func (m make) findInstalledFile(parentDir, filename string) (string, error) {
	var filePaths []string

	if err := filepath.Walk(m.InstalledDir, func(path string, info os.FileInfo, err error) error {
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
