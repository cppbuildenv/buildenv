package cli

import (
	"buildenv/config"
	"buildenv/console"
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

func (v *verifyCmd) listen() (handled bool) {
	if !v.verify {
		return false
	}

	var buildEnvConf config.BuildEnvConf
	if err := buildEnvConf.Verify(true, buildType.buildType); err != nil {
		platformName := strings.TrimSuffix(buildEnvConf.Platform, ".json")
		fmt.Printf(console.PlatformSelectedFailed, platformName, err)
		os.Exit(1)
	}

	// Silent mode called from buildenv.cmake
	if !silent.silent {
		platformName := strings.TrimSuffix(buildEnvConf.Platform, ".json")
		fmt.Printf(console.PlatformSelected, platformName)
	}

	return true
}
