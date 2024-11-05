package cli

import (
	"buildenv/config"
	"buildenv/console"
	"flag"
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
	gui            = newGUICmd(console.PlatformCallbacks)
	version        = newVersionCmd()
	createPlatform = newCreatePlatformCmd()
	selectPlatform = newSelectPlatformCmd(config.PlatformsDir, console.PlatformCallbacks)
	verify         = newVerifyCmd()
)
var commands = []reisterable{
	gui,
	version,
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
