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

const envTitle = "# buildenv runtime path (added by buildenv)"

func UpdateRunPath(runPath string) error {
	var homeDir = homeDir()
	var envFile string

	// Determine the environment file based on the OS.
	if runtime.GOOS == "linux" {
		envFile = ".profile"
	} else if runtime.GOOS == "darwin" {
		envFile = ".zshrc"
	} else {
		return fmt.Errorf("unsupported os: %s", runtime.GOOS)
	}

	envPath := filepath.Join(homeDir, envFile)

	// Open the file and read its content.
	file, err := os.OpenFile(envPath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return fmt.Errorf("cannot open %s file, err: %w", envFile, err)
	}
	defer file.Close()

	var lines []string
	var found bool

	// Scan through the file line by line, write existing lines to new lines,
	// except for the buildenv section.
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// If we find the # buildenv line, mark it and skip the next line.
		if strings.Contains(line, envTitle) {
			found = true
			continue // Skip the next line (the export PATH=...)
		}

		// If the buildenv section has been found, don't add the old export PATH line.
		if found && strings.HasPrefix(line, "export PATH=") {
			continue // Skip the export PATH line that follows the # buildenv line.
		}

		// Add the none empty line to the new content.
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}

	// Append buildenv section, it would be a new second or a replaced a second.
	lines = append(lines, "")
	lines = append(lines, envTitle)
	lines = append(lines, fmt.Sprintf("export PATH=%s", Join(runPath, "$PATH")))

	// Rewind the file and overwrite it with the updated content.
	file.Truncate(0) // Clear the content of the file.
	file.Seek(0, 0)  // Reset the file pointer to the beginning.

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

// Join joins the paths with the OS-specific path separator.
func Join(paths ...string) string {
	separator := string(string(os.PathListSeparator))
	return strings.Join(paths, separator)
}
