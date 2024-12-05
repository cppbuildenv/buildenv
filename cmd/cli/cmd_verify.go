package cli

import (
	"buildenv/config"
	"flag"
	"fmt"
	"strings"
)

func newVerifyCmd() *verifyCmd {
	return &verifyCmd{}
}

type verifyCmd struct {
	verify bool
}

func (v *verifyCmd) register() {
	flag.BoolVar(&v.verify, "verify", false, "check and repair cross build envronment for selected platform.")
}

func (v *verifyCmd) listen() (handled bool) {
	if !v.verify {
		return false
	}

	args := config.NewVerifyArgs(silent.silent, true, buildType.buildType)
	buildenv := config.NewBuildEnv(buildType.buildType)

	if err := buildenv.Verify(args); err != nil {
		platformName := strings.TrimSuffix(buildenv.Platform(), ".json")
		fmt.Print(config.PlatformSelectedFailed(platformName, err))
		return true
	}

	if !silent.silent {
		platformName := strings.TrimSuffix(buildenv.Platform(), ".json")
		fmt.Print(config.PlatformSelected(platformName))
	}

	return true
}
