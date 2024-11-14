package cli

import (
	"buildenv/console"
	"buildenv/pkg/env"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func newInstallCmd() *installCmd {
	return &installCmd{}
}

type installCmd struct {
	install bool
}

func (i *installCmd) register() {
	flag.BoolVar(&i.install, "install", false, "install buildenv so that can use it everywhere")
}

func (c *installCmd) listen() (handled bool) {
	if !c.install {
		return false
	}

	exePath, err := os.Executable()
	if err != nil {
		fmt.Print(console.InstallFailed(err))
		os.Exit(1)
	}

	if err := env.UpdateRunPath(filepath.Dir(exePath)); err != nil {
		fmt.Print(console.InstallFailed(err))
		os.Exit(1)
	}

	fmt.Print(console.InstallSuccess())
	return true
}
