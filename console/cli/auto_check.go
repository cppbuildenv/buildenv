package cli

import (
	"buildenv/config"
	"flag"
	"fmt"
)

func newAutoCheckCmd() *autoCheckCmd {
	return &autoCheckCmd{}
}

type autoCheckCmd struct {
	autoCheck bool
}

func (a *autoCheckCmd) register() {
	flag.BoolVar(&a.autoCheck, "a", false, "auto check buildenv")
	flag.BoolVar(&a.autoCheck, "autocheck", false, "auto check buildenv")
}

func (a *autoCheckCmd) listen() (handled bool) {
	if !a.autoCheck {
		return false
	}

	var buildEnvConf config.BuildEnvConf
	if err := buildEnvConf.Verify(); err != nil {
		fmt.Printf("[✘] ---- buildenv setup failed: %s", err)
	}

	return true
}
