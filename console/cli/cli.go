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
	force   forceCmd
	uimode  interactiveCmd
	version versionCmd
	create  createCmd
)
var commands = []reisterable{
	&force,
	&uimode,
	&version,
	&create,
}

// Listen listen commands input
func Listen() bool {
	// Read cmdName via flag
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
