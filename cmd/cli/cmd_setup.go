package cli

import (
	"buildenv/config"
	"flag"
)

func newSetupCmd() *setupCmd {
	return &setupCmd{}
}

type setupCmd struct {
	setup  bool
	silent bool
}

func (s *setupCmd) register() {
	flag.BoolVar(&s.setup, "setup", false, "setup cross build envronment for selected platform.")
	flag.BoolVar(&s.silent, "silent", false, "run buildenv no output, it's used with -setup.")
}

func (s *setupCmd) listen() (handled bool) {
	if !s.setup {
		return false
	}

	request := config.NewSetupArgs(s.silent, true, true).SetBuildType(buildType.buildType)
	buildenv := config.NewBuildEnv().SetBuildType(buildType.buildType)

	if err := buildenv.Setup(request); err != nil {
		config.PrintError(err, "failed to setup buildenv.")
		return true
	}

	if !s.silent {
		config.PrintSuccess("buildenv is ready for project: %s.", buildenv.ProjectName)
	}

	return true
}
