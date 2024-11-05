package env

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

const envTitle = "# buildenv (added by install)"

func UpdateRunPath(runPath string) error {
	var homeDir = homeDir()
	var envFile string

	// Determine the environment file based on the OS
	if runtime.GOOS == "linux" {
		envFile = ".profile"
	} else if runtime.GOOS == "darwin" {
		envFile = ".zshrc"
	} else {
		return fmt.Errorf("unsupported os: %s", runtime.GOOS)
	}

	envFilePath := filepath.Join(homeDir, envFile)

	// Open the file and read its content
	file, err := os.OpenFile(envFilePath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return fmt.Errorf("cannot open %s file, err: %w", envFile, err)
	}
	defer file.Close()

	var lines []string
	var foundBuildEnv bool
	scanner := bufio.NewScanner(file)

	// Scan through the file line by line
	for scanner.Scan() {
		line := scanner.Text()

		// If we find the # buildenv line, mark it and skip the next line
		if strings.Contains(line, envTitle) {
			foundBuildEnv = true
			continue // Skip the next line (the export PATH=...)
		}

		// If the buildenv section has been found, don't add the old export PATH line
		if foundBuildEnv && strings.HasPrefix(line, "export PATH=") {
			continue // Skip the export PATH line that follows the # buildenv line
		}

		// Add the line to the new content
		lines = append(lines, line)
	}

	// If we have already processed the file and found the section, we now append the new one.
	lines = append(lines, "")
	lines = append(lines, envTitle)
	lines = append(lines, fmt.Sprintf("export PATH=%s:$PATH", runPath))

	// Rewind the file and overwrite it with the updated content
	file.Truncate(0) // Clear the content of the file
	file.Seek(0, 0)  // Reset the file pointer to the beginning

	// Write the updated lines back to the file
	for _, line := range lines {
		if _, err := file.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write to %s, err: %w", envFile, err)
		}
	}

	return nil
}

func homeDir() string {
	usr, err := user.Current()
	if err != nil {
		panic("cannot get current user" + err.Error())
	}

	return usr.HomeDir
}
