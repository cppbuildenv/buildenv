package cli

import (
	"buildenv/config"
	"flag"
	"fmt"
	"os"
)

func handleSelect(callbacks config.BuildEnvCallbacks) {
	var (
		platform string
		project  string
	)

	cmd := flag.NewFlagSet("select", flag.ExitOnError)
	cmd.StringVar(&platform, "platform", "", "")
	cmd.StringVar(&project, "project", "", "")

	cmd.Usage = func() {
		fmt.Print("Usage: buildenv select [options]\n\n")
		fmt.Println("options:")
		cmd.PrintDefaults()
	}

	// Check if the --platform or --project is specified.
	if len(os.Args) < 3 {
		fmt.Println("Error: --platform or --project must be specified.")
		cmd.Usage()
		os.Exit(1)
	}

	cmd.Parse(os.Args[2:])

	if platform != "" {
		selectPlatform(platform, callbacks)
	} else if project != "" {
		selectProject(project, callbacks)
	}
}

func selectPlatform(platformName string, callbacks config.BuildEnvCallbacks) {
	if err := callbacks.OnSelectPlatform(platformName); err != nil {
		config.PrintError(err, "failed to select platform: %s.", platformName)
		os.Exit(1)
	}

	config.PrintSuccess("current platform: %s.", platformName)
}

func selectProject(projectName string, callbacks config.BuildEnvCallbacks) {
	if err := callbacks.OnSelectProject(projectName); err != nil {
		config.PrintError(err, "failed to select project: %s.", projectName)
		os.Exit(1)
	}
	config.PrintSuccess("buildenv is ready for project: %s.", projectName)
}
