package cli

import (
	"buildenv/console"
	"flag"
	"os"
	"runtime"
)

type reisterable interface {
	register()
}

type responsible interface {
	reisterable
	listen() (handled bool)
}

var (
	silent         = newSilentCmd()
	buildType      = newBuildTypeCmd()
	ui             = newUICmd(console.PlatformCallbacks)
	version        = newVersionCmd()
	sync           = newSyncCmd()
	createPlatform = newCreatePlatformCmd()
	selectPlatform = newSelectPlatformCmd(console.PlatformCallbacks)
	verify         = newVerifyCmd()
)
var commands = []reisterable{
	silent,
	buildType,
	ui,
	version,
	sync,
	createPlatform,
	selectPlatform,
	verify,
}

// Listen listen commands input
func Listen() bool {
	// `install` is supported in unix like system only.
	if runtime.GOOS == "linux" {
		install := newInstallCmd()
		commands = append(commands, install)
	}

	// Read command with flag
	for i := 0; i < len(commands); i++ {
		commands[i].register()
	}
	flag.Parse()

	// Handle commands
	for i := 0; i < len(commands); i++ {
		if cmd, ok := commands[i].(responsible); ok {
			if cmd.listen() {
				return true
			}
		}
	}

	return false
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	return !os.IsNotExist(err)
}
