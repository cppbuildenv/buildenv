package cli

import (
	"buildenv/config"
	"flag"
	"fmt"
	"os"
)

func handleInitialize(callbacks config.BuildEnvCallbacks) {
	var (
		url    string
		branch string
	)

	cmd := flag.NewFlagSet("init", flag.ExitOnError)
	cmd.StringVar(&url, "url", "", "conf repo url")
	cmd.StringVar(&branch, "branch", "master", "conf repo branch")

	cmd.Usage = func() {
		fmt.Print("Usage: buildenv init [options]\n\n")
		fmt.Println("options:")
		cmd.PrintDefaults()
	}

	cmd.Parse(os.Args[2:])

	output, err := callbacks.OnInitBuildEnv(url, branch)
	if err != nil {
		if url == "" {
			config.PrintError(err, "failed to init buildenv.")
		} else {
			config.PrintError(err, "failed to init buildenv with %s/%s.", url, branch)
		}
		return
	}

	fmt.Println(output)
	config.PrintSuccess("init buildenv successfully.")
}
