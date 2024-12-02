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
	flag.BoolVar(&v.verify, "verify", false, "check and repair for current selected platform.")
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
		os.Exit(1)
	}

	platformName := strings.TrimSuffix(buildenv.Platform(), ".json")
	fmt.Print(config.PlatformSelected(platformName))

	return true
}
