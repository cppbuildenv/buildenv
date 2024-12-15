package cli

import (
	"bufio"
	"buildenv/config"
	"buildenv/pkg/io"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func newUninstallCmd() *uninstallCmd {
	return &uninstallCmd{}
}

type uninstallCmd struct {
	uninstall string
}

func (u *uninstallCmd) register() {
	flag.StringVar(&u.uninstall, "uninstall", "", "uninstall a 3rd party port.")
}

func (u *uninstallCmd) listen() (handled bool) {
	if strings.TrimSpace(u.uninstall) == "" {
		return false
	}

	args := config.NewVerifyArgs(false, false, buildType.buildType)
	buildenv := config.NewBuildEnv(buildType.buildType)
	if err := buildenv.Verify(args); err != nil {
		fmt.Print(config.UninstallFailed(u.uninstall, err))
		return true
	}

	// Check if port to uninstall is exists in project.
	index := slices.IndexFunc(buildenv.Project().Ports, func(item string) bool {
		// exact match
		if item == u.uninstall {
			return true
		}

		// name match and the name must be someone of the ports in the project.
		if strings.Split(item, "-")[0] == u.uninstall {
			return true
		}

		return false
	})

	// Get the port to uninstall.
	var portToUninstall string
	if index == -1 {
		if !strings.Contains(u.uninstall, "-") {
			fmt.Print(config.UninstallFailed(u.uninstall,
				fmt.Errorf("cannot determine the exact port, as %s is not included in the port list of the current project", u.uninstall)))
			return true
		}

		portToUninstall = u.uninstall
	} else {
		portToUninstall = buildenv.Project().Ports[index]
	}

	// Check if port is installed.
	platformBuildType := fmt.Sprintf("%s-%s", buildenv.Platform().Name, buildenv.BuildType())
	installInfoFile := filepath.Join(config.Dirs.InstalledRootDir, "buildenv", "info", portToUninstall+"-"+platformBuildType+".list")
	if !io.PathExists(installInfoFile) {
		fmt.Print(config.UninstallFailed(portToUninstall, fmt.Errorf("%s is not installed", portToUninstall)))
		return true
	}

	// Open install info file.
	file, err := os.OpenFile(installInfoFile, os.O_RDONLY, os.ModePerm)
	if err != nil {
		fmt.Print(config.UninstallFailed(portToUninstall, fmt.Errorf("cannot open install info file: %s", err)))
		return true
	}
	defer file.Close()

	// Read line by line to remove installed file.
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// CMake project may generate a checksum file after install,
		// it would be like "/home/phil/.cmake/packages/gflags/4fbe0d242b1c0f095b87a43a7aeaf0d6",
		// We'll try to remove it also.
		fileToRemove := line
		if !io.PathExists(line) {
			fileToRemove = filepath.Join(config.Dirs.InstalledRootDir, line)
		}

		if err := os.Remove(fileToRemove); err != nil {
			fmt.Print(config.UninstallFailed(portToUninstall, err))
			return true
		}

		// Try remove parent folder if it's empty.
		if err := u.removeParentRecursively(filepath.Dir(fileToRemove)); err != nil {
			fmt.Print(config.UninstallFailed(portToUninstall, err))
			return true
		}

		fmt.Printf("remove %s\n", fileToRemove)
	}

	// Remove generated cmake config if exist.
	portName := strings.Split(portToUninstall, "-")[0]
	installedDir := filepath.Join(config.Dirs.InstalledRootDir, platformBuildType)
	if err := os.RemoveAll(filepath.Join(installedDir, "lib", "cmake", portName)); err != nil {
		fmt.Print(config.UninstallFailed(portToUninstall, err))
		return true
	}

	// Remove install info file.
	if err := os.Remove(installInfoFile); err != nil {
		fmt.Print(config.UninstallFailed(portToUninstall, err))
		return true
	}

	// Try remove installed dir.
	if err := u.removeParentRecursively(filepath.Dir(installInfoFile)); err != nil {
		fmt.Print(config.UninstallFailed(portToUninstall, err))
		return true
	}

	fmt.Print(config.UninstallSuccessfully(portToUninstall))

	return true
}

func (u uninstallCmd) removeParentRecursively(path string) error {
	entities, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	// Empty folder, remove it.
	if len(entities) == 0 {
		if err := os.RemoveAll(path); err != nil {
			return err
		}

		// Remove parent folder if it's empty.
		if err := u.removeParentRecursively(filepath.Dir(path)); err != nil {
			return err
		}

		return nil
	}

	return nil
}
