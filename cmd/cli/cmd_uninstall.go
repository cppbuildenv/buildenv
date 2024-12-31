package cli

import (
	"bufio"
	"buildenv/buildsystem"
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

	request := config.NewVerifyRequest(false, false, false).SetBuildType(buildType.buildType)
	buildenv := config.NewBuildEnv().SetBuildType(buildType.buildType)
	if err := buildenv.Verify(request); err != nil {
		config.PrintError(err, "%s uninstall failed.", u.uninstall)
		return true
	}

	// Check if port to uninstall is exists in project.
	index := slices.IndexFunc(buildenv.Project().Ports, func(item string) bool {
		// exact match
		if item == u.uninstall {
			return true
		}

		// name match and the name must be someone of the ports in the project.
		if strings.Split(item, "@")[0] == u.uninstall {
			return true
		}

		return false
	})

	// Get the port to uninstall.
	var portToUninstall string
	if index == -1 {
		if !strings.Contains(u.uninstall, "@") {
			config.PrintError(fmt.Errorf("cannot determine the exact port, because %s is not included in the port list of the current project", u.uninstall),
				"%s uninstall failed.", u.uninstall)
			return true
		}

		portToUninstall = u.uninstall
	} else {
		portToUninstall = buildenv.Project().Ports[index]
	}

	// Uninstall port.
	if err := u.uninstallPort(buildenv, portToUninstall, recursive.recursive); err != nil {
		config.PrintError(err, "%s uninstall failed.", u.uninstall)
		return true
	}

	config.PrintSuccess("%s uninstall successfully.", portToUninstall)

	return true
}

func (u uninstallCmd) uninstallPort(ctx config.Context, portNameVersion string, recursively bool) error {
	// Check port is configured ok.
	var port config.Port
	portPath := filepath.Join(config.Dirs.PortsDir, portNameVersion+".json")
	if err := port.Init(ctx, portPath); err != nil {
		return err
	}
	if err := port.Verify(); err != nil {
		return err
	}

	// No config found, download and deploy it.
	if len(port.BuildConfigs) == 0 {
		return nil
	}

	// Find matched config.
	var matchedConfig *buildsystem.BuildConfig
	for _, config := range port.BuildConfigs {
		if port.MatchPattern(config.Pattern) {
			matchedConfig = &config
			break
		}
	}
	if matchedConfig == nil {
		return fmt.Errorf("no matching build_config found to build")
	}

	// Try to uninstall dependencies firstly.
	if recursively {
		for _, item := range matchedConfig.Depedencies {
			if strings.HasPrefix(item, port.Name) {
				return fmt.Errorf("port.dependencies contains circular dependency: %s", item)
			}

			// Check and verify dependency.
			var port config.Port
			portPath := filepath.Join(config.Dirs.PortsDir, item+".json")
			if err := port.Init(ctx, portPath); err != nil {
				return err
			}
			if err := port.Verify(); err != nil {
				return err
			}

			// Uninstall dependency.
			if err := u.uninstallPort(ctx, item, recursively); err != nil {
				return err
			}
		}
	}

	// Do uninstall port itself.
	if err := u.doUninsallPort(ctx, port.NameVersion()); err != nil {
		return err
	}

	return nil
}

func (u uninstallCmd) doUninsallPort(ctx config.Context, portNameVersion string) error {
	// Check if port is installed.
	platformBuildType := fmt.Sprintf("%s-%s", ctx.Platform().Name, ctx.BuildType())
	installInfoFile := filepath.Join(config.Dirs.InstalledRootDir, "buildenv", "info", portNameVersion+"-"+platformBuildType+".list")
	if !io.PathExists(installInfoFile) {
		return fmt.Errorf("%s is not installed", portNameVersion)
	}

	// Open install info file.
	file, err := os.OpenFile(installInfoFile, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("cannot open install info file: %s", err)
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
		if err := u.removeFiles(fileToRemove); err != nil {
			return fmt.Errorf("cannot remove file: %s", err)
		}

		// Try remove parent folder if it's empty.
		if err := u.removeFolderRecursively(filepath.Dir(fileToRemove)); err != nil {
			return fmt.Errorf("cannot remove parent folder: %s", err)
		}

		fmt.Printf("remove %s\n", fileToRemove)
	}

	// Remove generated cmake config if exist.
	portName := strings.Split(portNameVersion, "@")[0]
	installedDir := filepath.Join(config.Dirs.InstalledRootDir, platformBuildType)
	cmakeConfigDir := filepath.Join(installedDir, "lib", "cmake", portName)
	if err := os.RemoveAll(cmakeConfigDir); err != nil {
		return fmt.Errorf("cannot remove cmake config folder: %s", err)
	}
	if err := u.removeFolderRecursively(filepath.Dir(cmakeConfigDir)); err != nil {
		return fmt.Errorf("cannot clean cmake config folder: %s", err)
	}

	// Remove install info file.
	if err := os.Remove(installInfoFile); err != nil {
		return fmt.Errorf("cannot remove install info file: %s", err)
	}

	// Try to clean installed dir.
	if err := u.removeFolderRecursively(filepath.Dir(installInfoFile)); err != nil {
		return fmt.Errorf("cannot remove parent folder: %s", err)
	}

	return nil
}

// removeFiles remove files and all related shared libraries.
func (u uninstallCmd) removeFiles(path string) error {
	if !strings.Contains(path, "so") {
		return os.Remove(path)
	}

	index := strings.Index(path, ".so")
	matches, err := filepath.Glob(path[:index] + ".so*")
	if err != nil {
		return err
	}

	for _, item := range matches {
		if err := os.Remove(item); err != nil {
			return err
		}
	}

	return nil
}

func (u uninstallCmd) removeFolderRecursively(path string) error {
	// Not exists, skip.
	if !io.PathExists(path) {
		return nil
	}

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
		if err := u.removeFolderRecursively(filepath.Dir(path)); err != nil {
			return err
		}

		return nil
	}

	return nil
}
