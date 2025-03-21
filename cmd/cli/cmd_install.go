package cli

import (
	"buildenv/config"
	"buildenv/pkg/fileio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func handleInstall(callbacks config.BuildEnvCallbacks) {
	var (
		buildType string
		dev       bool
	)

	cmd := flag.NewFlagSet("install", flag.ExitOnError)
	cmd.StringVar(&buildType, "build_type", "Release", "build type, for example: Release, Debug, etc.")
	cmd.BoolVar(&dev, "dev", false, "install a dev third-party.")

	cmd.Usage = func() {
		fmt.Print("Usage: buildenv install <name@version|name>\n\n")
	}

	// Check if the <name@value|name> is specified.
	if len(os.Args) < 3 {
		fmt.Println("Error: The <name@value|name> must be specified.")
		cmd.Usage()
		os.Exit(1)
	}

	cmd.Parse(os.Args[3:])
	nameVersion := os.Args[2]

	// Make sure toolchain, rootfs and tools are prepared.
	args := config.NewSetupArgs(false, true, false).SetBuildType(buildType)
	buildEnvPath := filepath.Join(config.Dirs.WorkspaceDir, "buildenv.json")

	buildenv := config.NewBuildEnv().SetBuildType(buildType)
	if err := buildenv.Init(buildEnvPath); err != nil {
		config.PrintError(err, "failed to init buildenv %s: %s.", nameVersion, err)
		return
	}
	if err := buildenv.Setup(args); err != nil {
		config.PrintError(err, "install %s failed.", nameVersion)
		return
	}

	// Exact check if port to install is exists.
	if strings.Count(nameVersion, "@") > 0 {
		parts := strings.Split(nameVersion, "@")
		portPaths := filepath.Join(config.Dirs.PortsDir, parts[0], parts[1]+".json")
		if !fileio.PathExists(portPaths) {
			config.PrintError(fmt.Errorf("port %s is not found", nameVersion), "%s install failed.", nameVersion)
			return
		}
	} else {
		// Check if port to install is exists in project.
		index := slices.IndexFunc(buildenv.Project().Ports, func(item string) bool {
			return strings.Split(item, "@")[0] == nameVersion
		})
		if index == -1 {
			config.PrintError(fmt.Errorf("port %s is not found", nameVersion), "%s install failed.", nameVersion)
			return
		}
	}

	// Install the port.
	var port config.Port
	port.AsDev = dev
	if err := port.Init(buildenv, nameVersion); err != nil {
		config.PrintError(err, "install %s failed.", nameVersion)
		return
	}
	if err := port.Install(false); err != nil {
		config.PrintError(err, "install %s failed.", nameVersion)
		return
	}

	config.PrintSuccess("install %s successfully.", nameVersion)
}
