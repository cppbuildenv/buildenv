package cli

import (
	"buildenv/config"
	"flag"
	"fmt"
	"os"
)

func handleSetup(callbacks config.BuildEnvCallbacks) {
	var (
		silent    bool
		buildType string
	)

	cmd := flag.NewFlagSet("setup", flag.ExitOnError)
	cmd.BoolVar(&silent, "silent", false, "run in silent mode, no output log.")
	cmd.StringVar(&buildType, "build_type", "Release", "build type, for example: Release, Debug, etc.")

	cmd.Usage = func() {
		fmt.Print("Usage: buildenv setup [options]\n\n")
		fmt.Println("options:")
		cmd.PrintDefaults()
	}

	cmd.Parse(os.Args[2:])
	args := config.NewSetupArgs(silent, true, true).SetBuildType(buildType)
	buildenv := config.NewBuildEnv().SetBuildType(buildType)

	if err := buildenv.Setup(args); err != nil {
		config.PrintError(err, "failed to setup buildenv.")
		return
	}

	if !silent {
		config.PrintSuccess("buildenv is ready for project: %s.", buildenv.ProjectName)
	}
}
