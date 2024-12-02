package cli

import (
	"buildenv/config"
	"buildenv/pkg/io"
	"flag"
	"fmt"
	"path/filepath"
	"strings"
)

func newInstallCmd() *installCmd {
	return &installCmd{}
}

type installCmd struct {
	install string
}

func (i *installCmd) register() {
	flag.StringVar(&i.install, "install", "", "configre, build and install a port")
}

func (i *installCmd) listen() (handled bool) {
	if strings.TrimSpace(i.install) == "" {
		return false
	}

	// Check port config is exists.
	portPath := filepath.Join(config.Dirs.PortDir, i.install+".json")
	if !io.PathExists(portPath) {
		fmt.Print(config.InstallFailed(i.install, fmt.Errorf("port config is not exists")))
		return true
	}

	// Configure, build and install specified port.
	args := config.NewVerifyArgs(silent.silent, true, buildType.buildType)
	args.SetVerifyPort(i.install)
	buildenv := config.NewBuildEnv(buildType.buildType)
	if err := buildenv.Verify(args); err != nil {
		fmt.Print(config.InstallFailed(i.install, err))
		return true
	}

	return true
}
