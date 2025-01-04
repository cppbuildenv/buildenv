package cli

import (
	"buildenv/config"
	"buildenv/pkg/fileio"
	"flag"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

func newInstallCmd() *installCmd {
	return &installCmd{}
}

type installCmd struct {
	install string
}

func (i *installCmd) register() {
	flag.StringVar(&i.install, "install", "", "clone, configre, build and install a 3rd party port.")
}

func (i *installCmd) listen() (handled bool) {
	if strings.TrimSpace(i.install) == "" {
		return false
	}

	// Make sure toolchain, rootfs and tools are prepared.
	request := config.NewVerifyRequest(verify.silent, true, false).
		SetBuildType(buildType.buildType)
	buildEnvPath := filepath.Join(config.Dirs.WorkspaceDir, "buildenv.json")

	buildenv := config.NewBuildEnv().SetBuildType(buildType.buildType)
	if err := buildenv.Init(buildEnvPath); err != nil {
		config.PrintError(err, "failed to init buildenv before install %s: %s.", i.install, err)
		return true
	}
	if err := buildenv.Verify(request); err != nil {
		config.PrintError(err, "%s install failed.", i.install)
		return true
	}

	// Exact check if port to install is exists.
	var portToInstall string
	parts := strings.Split(i.install, "@")
	if len(parts) == 2 {
		portPaths := filepath.Join(config.Dirs.PortsDir, parts[0], parts[1]+".json")
		if !fileio.PathExists(portPaths) {
			config.PrintError(fmt.Errorf("port %s is not found", i.install), "%s install failed.", i.install)
			return true
		}

		portToInstall = filepath.Join(parts[0], parts[1])
	} else {
		// Check if port to install is exists in project.
		index := slices.IndexFunc(buildenv.Project().Ports, func(item string) bool {
			return strings.Split(item, "@")[0] == i.install
		})
		if index == -1 {
			config.PrintError(fmt.Errorf("port %s is not found", i.install), "%s install failed.", i.install)
			return true
		}

		parts := strings.Split(buildenv.Project().Ports[index], "@")
		portToInstall = filepath.Join(parts[0], parts[1])
	}

	// Install the port.
	var port config.Port
	portPath := filepath.Join(config.Dirs.PortsDir, portToInstall+".json")
	if err := port.Init(buildenv, portPath); err != nil {
		config.PrintError(err, "%s install failed.", i.install)
		return true
	}
	if err := port.Verify(); err != nil {
		config.PrintError(err, "%s install failed.", i.install)
		return true
	}
	if err := port.Install(verify.silent); err != nil {
		config.PrintError(err, "%s install failed.", i.install)
		return true
	}

	config.PrintSuccess("%s install successfully.", portToInstall)
	return true
}
