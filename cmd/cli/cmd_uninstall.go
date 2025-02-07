package cli

import (
	"bufio"
	"buildenv/buildsystem"
	"buildenv/config"
	"buildenv/pkg/fileio"
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
	recursive bool
	purge     bool
}

func (u *uninstallCmd) register() {
	flag.StringVar(&u.uninstall, "uninstall", "", "uninstall a 3rd party port.")
	flag.BoolVar(&u.recursive, "recursive", false, "uninstall dependencies also, it works with -uninstall.")
	flag.BoolVar(&u.purge, "purge", false, "remove installed files after uninstall, it works with -uninstall.")
}

func (u *uninstallCmd) listen() (handled bool) {
	if strings.TrimSpace(u.uninstall) == "" {
		return false
	}

	args := config.NewSetupArgs(false, false, false).SetBuildType(buildType.buildType)
	buildenv := config.NewBuildEnv().SetBuildType(buildType.buildType)
	if err := buildenv.Setup(args); err != nil {
		config.PrintError(err, "%s uninstall failed.", u.uninstall)
		return true
	}

	// Check if install port as dev.
	asDev := strings.HasSuffix(u.uninstall, "@dev")
	u.uninstall = strings.TrimSuffix(u.uninstall, "@dev")

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
			config.PrintError(fmt.Errorf("cannot determine the exact port, "+
				"because %s is not included in the port list of the current project", u.uninstall),
				"%s uninstall failed.", u.uninstall)
			return true
		}

		portToUninstall = u.uninstall
	} else {
		portToUninstall = buildenv.Project().Ports[index]
	}

	// Uninstall port.
	if err := u.uninstallPort(buildenv, portToUninstall, asDev); err != nil {
		config.PrintError(err, "%s uninstall failed.", u.uninstall)
		return true
	}

	config.PrintSuccess("%s uninstall successfully.", portToUninstall)

	return true
}

func (u uninstallCmd) uninstallPort(ctx config.Context, nameVersion string, asDev bool) error {
	// Check port is configured ok.
	var port config.Port
	port.AsDev = asDev
	if err := port.Init(ctx, nameVersion); err != nil {
		return err
	}
	if err := port.Validate(); err != nil {
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
	if u.recursive {
		uninstall := func(nameVersion string, asDev bool) error {
			if strings.HasPrefix(nameVersion, port.Name) {
				return fmt.Errorf("%s's dependencies contains circular dependency: %s",
					port.NameVersion(), nameVersion)
			}

			// Check and validate dependency.
			var port config.Port
			port.AsDev = asDev
			port.AsSubDep = true
			if err := port.Init(ctx, nameVersion); err != nil {
				return err
			}
			if err := port.Validate(); err != nil {
				return err
			}

			// Uninstall dependency.
			if err := u.uninstallPort(ctx, nameVersion, asDev); err != nil {
				return err
			}

			return nil
		}

		for _, nameVersion := range matchedConfig.Depedencies {
			if err := uninstall(nameVersion, false); err != nil {
				return err
			}
		}
		for _, nameVersion := range matchedConfig.DevDepedencies {
			if err := uninstall(nameVersion, true); err != nil {
				return err
			}
		}
	}

	// Do uninstall port itself.
	if err := u.doUninsallPort(ctx, port); err != nil {
		return err
	}

	// Remove package files if purge option is specified.
	if u.purge {
		var folderName string
		if port.AsDev {
			folderName = port.NameVersion()
		} else {
			folderName = fmt.Sprintf("%s-%s-%s-%s",
				port.NameVersion(),
				ctx.Platform().Name,
				ctx.Project().Name,
				ctx.BuildType())
		}

		// Remove port's package files.
		packageDir := filepath.Join(config.Dirs.WorkspaceDir, "packages", folderName)
		if err := os.RemoveAll(packageDir); err != nil {
			return fmt.Errorf("cannot remove package files: %s", err)
		}

		// Try remove parent folder if it's empty.
		if err := fileio.RemoveFolderRecursively(filepath.Dir(packageDir)); err != nil {
			return fmt.Errorf("cannot remove parent folder: %s", err)
		}
	}

	return nil
}

func (u uninstallCmd) doUninsallPort(ctx config.Context, port config.Port) error {
	// Check if port is installed.
	var stateFileName string
	if port.AsDev {
		stateFileName = fmt.Sprintf("%s-dev.list", port.NameVersion())
	} else {
		stateFileName = fmt.Sprintf("%s-%s-%s-%s.list", port.NameVersion(), ctx.Platform().Name, ctx.Project().Name, ctx.BuildType())
	}
	stateFilePath := filepath.Join(config.Dirs.WorkspaceDir, "installed", "buildenv", "info", stateFileName)
	if !fileio.PathExists(stateFilePath) {
		return fmt.Errorf("%s is not installed", port.NameVersion())
	}

	// Open install info file.
	file, err := os.OpenFile(stateFilePath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("cannot open install info file: %s", err)
	}
	defer file.Close()

	platformProject := fmt.Sprintf("%s-%s-%s", ctx.Platform().Name, ctx.Project().Name, ctx.BuildType())

	// Read line by line to remove installed file.
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// CMake project may generate a checksum file after install,
		// it would be like "/home/phil/.cmake/packages/gflags/4fbe0d242b1c0f095b87a43a7aeaf0d6",
		// We'll try to remove it also.
		fileToRemove := line
		if !fileio.PathExists(line) {
			fileToRemove = filepath.Join(config.Dirs.WorkspaceDir, "installed", line)
		}
		if err := u.removeFiles(fileToRemove); err != nil {
			return fmt.Errorf("cannot remove file: %s", err)
		}

		// Try remove parent folder if it's empty.
		if err := fileio.RemoveFolderRecursively(filepath.Dir(fileToRemove)); err != nil {
			return fmt.Errorf("cannot remove parent folder: %s", err)
		}

		fmt.Printf("remove %s\n", fileToRemove)
	}

	// Remove generated cmake config if exist.
	portName := strings.Split(port.NameVersion(), "@")[0]
	cmakeConfigDir := filepath.Join(config.Dirs.InstalledDir, platformProject, "lib", "cmake", portName)
	if err := os.RemoveAll(cmakeConfigDir); err != nil {
		return fmt.Errorf("cannot remove cmake config folder: %s", err)
	}
	if err := fileio.RemoveFolderRecursively(filepath.Dir(cmakeConfigDir)); err != nil {
		return fmt.Errorf("cannot clean cmake config folder: %s", err)
	}

	// Remove install info file.
	if err := os.Remove(stateFilePath); err != nil {
		return fmt.Errorf("cannot remove install info file: %s", err)
	}

	// Try to clean installed dir.
	if err := fileio.RemoveFolderRecursively(filepath.Dir(stateFilePath)); err != nil {
		return fmt.Errorf("cannot remove parent folder: %s", err)
	}

	return nil
}

// removeFiles remove files and all related shared libraries.
func (u uninstallCmd) removeFiles(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	if !strings.Contains(path, "so") {
		return os.Remove(path)
	}

	index := strings.Index(path, ".so")
	if index == -1 {
		return os.Remove(path)
	}

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
