package cli

import (
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
	force          = newForceCmd()
	interactive    = newInteractiveCmd()
	version        = newVersionCmd()
	createPlatform = newCreatePlatformCmd()
	selectPlatform = newSelectPlatformCmd("", func(fullpath string) {

	})
)
var commands = []reisterable{
	force,
	interactive,
	version,
	createPlatform,
	selectPlatform,
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
