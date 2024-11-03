package cli

import (
	"buildenv/config"
	"buildenv/console"
	"flag"
)

type reisterable interface {
	register()
}

type responsible interface {
	reisterable
	listen() (handled bool)
}

var (
	interactive    = newInteractiveCmd(console.PlatformCallbacks)
	version        = newVersionCmd()
	createPlatform = newCreatePlatformCmd()
	selectPlatform = newSelectPlatformCmd(config.PlatformDir, console.PlatformCallbacks)
	autoCheck      = newAutoCheckCmd()
)
var commands = []reisterable{
	interactive,
	version,
	createPlatform,
	selectPlatform,
	autoCheck,
}

// Listen listen commands input
func Listen() bool {
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
