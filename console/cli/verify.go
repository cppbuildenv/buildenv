package cli

import (
	"buildenv/config"
	"flag"
	"fmt"
	"os"
	"strings"
)

func newVerifyCmd() *verifyCmd {
	return &verifyCmd{}
}

type verifyCmd struct {
	verify bool
}

func (v *verifyCmd) register() {
	flag.BoolVar(&v.verify, "verify", false, "verify buildenv")
}

func (a *verifyCmd) listen() (handled bool) {
	if !a.verify {
		return false
	}

	var buildEnvConf config.BuildEnvConf
	if err := buildEnvConf.Verify(true); err != nil {
		platformName := strings.TrimSuffix(buildEnvConf.Platform, ".json")
		fmt.Printf("[✘] ---- buildenv setup failed: %s: %s\n\n", platformName, err)
		os.Exit(1)
	}

	platformName := strings.TrimSuffix(buildEnvConf.Platform, ".json")
	fmt.Printf("[✔] ==== buildenv setup successfully: %s\n", platformName)

	return true
}
