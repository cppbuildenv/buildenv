package cli

import (
	"buildenv/config"
	"buildenv/pkg/env"
	"flag"
	"os"
	"path/filepath"
)

func newIntegrateCmd() *integrateCmd {
	return &integrateCmd{}
}

type integrateCmd struct {
	integrate bool
}

func (i *integrateCmd) register() {
	flag.BoolVar(&i.integrate, "integrate", false, "integrate buildenv so can use it everywhere.")
}

func (c *integrateCmd) listen() (handled bool) {
	if !c.integrate {
		return false
	}

	exePath, err := os.Executable()
	if err != nil {
		config.PrintError(err, "buildenv integrate failed.")
		os.Exit(1)
	}

	if err := env.UpdateRunPath(filepath.Dir(exePath)); err != nil {
		config.PrintError(err, "buildenv integrate failed.")
		os.Exit(1)
	}

	config.PrintSuccess("buildenv is integrated.")
	return true
}
