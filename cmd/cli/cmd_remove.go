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

func newRemoveCmd() *removeCmd {
	return &removeCmd{}
}

type removeCmd struct {
	remove      string
	purge       string
	recursive   bool
	portRemoved func(ctx config.Context, port config.Port) error
}

func (r *removeCmd) register() {
	flag.StringVar(&r.remove, "remove", "", "remove a third-party from installed dir, for example: glog@v0.6.0, "+
		" you can also call with '--dev' suffix to remove a dev third-pary.")
	flag.BoolVar(&r.recursive, "recursive", false, "remove a third-party with dependencies also, it works with --remove and --purge.")
}

func (r *removeCmd) listen() (handled bool) {
	var targetPort string
	if strings.TrimSpace(r.remove) != "" {
		targetPort = r.remove
	} else if strings.TrimSpace(r.purge) != "" {
		targetPort = r.purge
	}
	if targetPort == "" {
		return false
	}

	args := config.NewSetupArgs(false, false, false).SetBuildType(buildType.buildType)
	buildenv := config.NewBuildEnv().SetBuildType(buildType.buildType)
	if err := buildenv.Setup(args); err != nil {
		config.PrintError(err, "%s remove failed.", targetPort)
		return true
	}

	// Check if port to remove is exists in project.
	index := slices.IndexFunc(buildenv.Project().Ports, func(item string) bool {
		// exact match
		if item == targetPort {
			return true
		}

		// name match and the name must be someone of the ports in the project.
		if strings.Split(item, "@")[0] == targetPort {
			return true
		}

		return false
	})

	// Get the port to remove.
	var portToRemove string
	if index == -1 {
		if !strings.Contains(targetPort, "@") {
			config.PrintError(fmt.Errorf("cannot determine the exact port, "+
				"because %s is not included in the port list of the current project", r.remove),
				"%s remove failed.", r.remove)
			return true
		}

		portToRemove = targetPort
	} else {
		portToRemove = buildenv.Project().Ports[index]
	}

	// Remove port.
	if err := r.removePort(buildenv, portToRemove, dev.dev); err != nil {
		config.PrintError(err, "%s remove failed.", targetPort)
		return true
	}

	config.PrintSuccess("%s remove successfully.", portToRemove)
	return true
}

func (r removeCmd) removePort(ctx config.Context, nameVersion string, asDev bool) error {
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

	// Try to remove dependencies firstly.
	if r.recursive {
		remove := func(nameVersion string, asDev bool) error {
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

			// Remove dependency.
			if err := r.removePort(ctx, nameVersion, asDev); err != nil {
				return err
			}

			return nil
		}

		for _, nameVersion := range matchedConfig.Depedencies {
			if err := remove(nameVersion, false); err != nil {
				return err
			}
		}
		for _, nameVersion := range matchedConfig.DevDepedencies {
			if err := remove(nameVersion, true); err != nil {
				return err
			}
		}
	}

	// Do remove port itself.
	if err := r.doRemovePort(ctx, port); err != nil {
		return err
	}

	// PurgeCmd would listen to this callback to remove port from package.
	if r.portRemoved != nil {
		if err := r.portRemoved(ctx, port); err != nil {
			return err
		}
	}

	return nil
}

func (r removeCmd) doRemovePort(ctx config.Context, port config.Port) error {
	// Check if port is installed.
	var stateFileName string
	if port.AsDev {
		stateFileName = fmt.Sprintf("%s^dev.list", port.NameVersion())
	} else {
		stateFileName = fmt.Sprintf("%s^%s^%s^%s.list", port.NameVersion(), ctx.Platform().Name, ctx.Project().Name, ctx.BuildType())
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

	platformProject := fmt.Sprintf("%s^%s^%s", ctx.Platform().Name, ctx.Project().Name, ctx.BuildType())

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
		if err := r.removeFiles(fileToRemove); err != nil {
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
func (r removeCmd) removeFiles(path string) error {
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
