package cli

import (
	"buildenv/config"
	"flag"
	"fmt"
)

func newVerifyCmd() *verifyCmd {
	return &verifyCmd{}
}

type verifyCmd struct {
	verify bool
	silent bool
}

func (v *verifyCmd) register() {
	flag.BoolVar(&v.verify, "verify", false, "check and repair cross build envronment for selected platform.")
	flag.BoolVar(&v.silent, "silent", false, "run buildenv no output, it's used with -verify.")
}

func (v *verifyCmd) listen() (handled bool) {
	if !v.verify {
		return false
	}

	args := config.NewVerifyArgs(v.silent, true, buildType.buildType)
	buildenv := config.NewBuildEnv().SetBuildType(buildType.buildType)

	if err := buildenv.Verify(args); err != nil {
		fmt.Print(config.PlatformSelectedFailed(buildenv.PlatformName, err))
		return true
	}

	if !v.silent {
		fmt.Print(config.ProjectSelected(buildenv.ProjectName))
	}

	return true
}
