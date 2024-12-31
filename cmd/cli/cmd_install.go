package cli

import (
	"buildenv/config"
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

	// Configure, build and install a port.
	verifyArgs := config.NewVerifyArgs(verify.silent, false).SetBuildType(buildType.buildType)
	buildenv := config.NewBuildEnv().SetBuildType(buildType.buildType)
	if err := buildenv.Verify(verifyArgs); err != nil {
		config.PrintError(err, "%s install failed.", i.install)
		return true
	}

	// Check if port to install is exists in project.
	index := slices.IndexFunc(buildenv.Project().Ports, func(item string) bool {
		// exact match
		if item == i.install {
			return true
		}

		// name match and the name must be someone of the ports in the project.
		if strings.Split(item, "@")[0] == i.install {
			return true
		}

		return false
	})
	if index == -1 {
		config.PrintError(fmt.Errorf("port %s is not found", i.install), "%s install failed.", i.install)
		return true
	}

	// Install the port.
	portToInstall := buildenv.Project().Ports[index]
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
	installArgs := config.NewVerifyArgs(verify.silent, true).SetBuildType(buildType.buildType)
	if err := port.CheckAndRepair(installArgs); err != nil {
		config.PrintError(err, "%s install failed.", i.install)
		return true
	}

	config.PrintSuccess("%s install successfully.", portToInstall)
	return true
}
