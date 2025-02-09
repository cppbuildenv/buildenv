package cli

import (
	"buildenv/config"
	"buildenv/pkg/env"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func handleIntegrate(callbacks config.BuildEnvCallbacks) {
	cmd := flag.NewFlagSet("integrate", flag.ExitOnError)

	cmd.Usage = func() {
		fmt.Print("Usage: buildenv integrate\n\n")
		fmt.Println("options:")
		cmd.PrintDefaults()
	}

	exePath, err := os.Executable()
	if err != nil {
		config.PrintError(err, "buildenv integrate failed.")
		os.Exit(1)
	}

	if err := env.UpdateRunPath(filepath.Dir(exePath)); err != nil {
		config.PrintError(err, "buildenv integrate failed.")
		os.Exit(1)
	}

	config.PrintSuccess("buildenv is integrated.")
}
