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
		panic(fmt.Sprintf(console.InstallFailed, err))
	}

	if err := env.UpdateRunPath(filepath.Dir(exePath)); err != nil {
		panic(fmt.Sprintf(console.InstallFailed, err))
	}

	fmt.Printf(console.InstallSuccess)
	return true
}
