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

func handleRemove(callbacks config.BuildEnvCallbacks) {
	var (
		buildType string
		recurse   bool
		purge     bool
		dev       bool
	)

	cmd := flag.NewFlagSet("remove", flag.ExitOnError)
	cmd.StringVar(&buildType, "build_type", "Release", "build type, for example: Release, Debug, etc.")
	cmd.BoolVar(&recurse, "recurse", false, "Remove a third-party with its dependencies also.")
	cmd.BoolVar(&purge, "purge", false, "Remove a third-party with its package also.")
	cmd.BoolVar(&dev, "dev", false, "Remove a dev third-party.")

	cmd.Usage = func() {
		fmt.Print("Usage: buildenv remove <name@value|name> [options]\n\n")
		fmt.Println("options:")
		cmd.PrintDefaults()
	}

	// Check if the <name@value|name> is specified.
	if len(os.Args) < 3 {
		fmt.Println("Error: The <name@value|name> must be specified.")
		cmd.Usage()
		os.Exit(1)
	}

	cmd.Parse(os.Args[3:])
	nameVersion := os.Args[2]

	args := config.NewSetupArgs(false, false, false).SetBuildType(buildType)
	buildenv := config.NewBuildEnv().SetBuildType(buildType)
	if err := buildenv.Setup(args); err != nil {
		config.PrintError(err, "%s remove failed.", nameVersion)
		os.Exit(1)
	}

	// Check if port to remove exists in project.
	index := slices.IndexFunc(buildenv.Project().Ports, func(item string) bool {
		// exact match
		if item == nameVersion {
			return true
		}

		// name match and the name must be someone of the ports in the project.
		if strings.Split(item, "@")[0] == nameVersion {
			return true
		}

		return false
	})

	// Get the port to remove.
	var portToRemove string
	if index == -1 {
		if !strings.Contains(nameVersion, "@") {
			config.PrintError(fmt.Errorf("cannot determine the exact port, "+
				"because %s is not included in the port list of the current project", nameVersion),
				"%s remove failed.", nameVersion)
			os.Exit(1)
		}

		portToRemove = nameVersion
	} else {
		portToRemove = buildenv.Project().Ports[index]
	}

	// Remove port.
	if err := removePort(buildenv, portToRemove, dev, purge, recurse); err != nil {
		config.PrintError(err, "%s remove failed.", nameVersion)
		os.Exit(1)
	}

	config.PrintSuccess("%s remove successfully.", portToRemove)
}

func removePort(ctx config.Context, nameVersion string, asDev, purge, recurse bool) error {
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
	if recurse {
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
			if err := removePort(ctx, nameVersion, asDev, purge, recurse); err != nil {
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
	if err := doRemovePort(ctx, port); err != nil {
		return err
	}

	// Remove port's package files.
	if purge {
		if err := removePackage(ctx, port); err != nil {
			return err
		}
	}

	return nil
}

func doRemovePort(ctx config.Context, port config.Port) error {
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
		if err := removeFiles(fileToRemove); err != nil {
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

func removePackage(ctx config.Context, port config.Port) error {
	var folderName string
	if port.AsDev {
		folderName = port.NameVersion()
	} else {
		folderName = fmt.Sprintf("%s^%s^%s^%s",
			port.NameVersion(),
			ctx.Platform().Name,
			ctx.Project().Name,
			ctx.BuildType(),
		)
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

	return nil
}

// removeFiles remove files and all related shared libraries.
func removeFiles(path string) error {
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
