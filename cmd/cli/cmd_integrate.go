package cli

import (
	"buildenv/config"
	"buildenv/pkg/env"
	"flag"
	"fmt"
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
		fmt.Print(config.IntegrateFailed(err))
		os.Exit(1)
	}

	if err := env.UpdateRunPath(filepath.Dir(exePath)); err != nil {
		fmt.Print(config.IntegrateFailed(err))
		os.Exit(1)
	}

	fmt.Print(config.IntegrateSuccessfully())
	return true
}
