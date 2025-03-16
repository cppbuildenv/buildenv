package cmd

import (
	"fmt"
	"strings"
)

// IsLibraryInstalled checks if a library is installed on the system.
func IsLibraryInstalled(libraryName string) (bool, error) {
	osType, err := getOSType()
	if err != nil {
		return false, err
	}

	switch osType {
	case "debian", "ubuntu":
		return checkDebianLibrary(libraryName)
	case "centos", "fedora", "rhel":
		return checkRedHatLibrary(libraryName)
	default:
		return false, fmt.Errorf("unsupported OS type: %s", osType)
	}
}

func getOSType() (string, error) {
	executor := NewExecutor("", "cat /etc/os-release")
	out, err := executor.ExecuteOutput()
	if err != nil {
		return "", fmt.Errorf("failed to read /etc/os-release: %v", err)
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ID=") {
			id := strings.TrimPrefix(line, "ID=")
			id = strings.Trim(id, `"`)
			return id, nil
		}
	}

	return "", fmt.Errorf("failed to determine OS type")
}

func checkDebianLibrary(libraryName string) (bool, error) {
	// Use dpkg -l to check if the library is installed.
	executor := NewExecutor("", "dpkg -l "+libraryName)
	out, err := executor.ExecuteOutput()
	if err != nil {
		// If not installed, dpkg -l will return exit status 1.
		return false, nil
	}

	// Check if the library is installed.
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ii") && strings.Contains(line, libraryName) {
			return true, nil
		}
	}

	return false, nil
}

func checkRedHatLibrary(libraryName string) (bool, error) {
	// Use rpm -q to check if the library is installed
	executor := NewExecutor("", "rpm -q "+libraryName)
	out, err := executor.ExecuteOutput()
	if err != nil {
		return false, fmt.Errorf("failed to run rpm -q: %v", err)
	}

	// Check if the library is installed.
	if !strings.Contains(string(out), "not installed") {
		return true, nil
	}

	return false, nil
}
