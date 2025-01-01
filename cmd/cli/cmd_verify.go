package cli

import (
	"buildenv/config"
	"flag"
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

	request := config.NewVerifyRequest(v.silent, true, true).SetBuildType(buildType.buildType)
	buildenv := config.NewBuildEnv().SetBuildType(buildType.buildType)

	if err := buildenv.Verify(request); err != nil {
		config.PrintError(err, "failed to verify buildenv.")
		return true
	}

	if !v.silent {
		config.PrintSuccess("buildenv is ready for project: %s.", buildenv.ProjectName)
	}

	return true
}
