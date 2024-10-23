package cmd

import (
	"flag"
)

// Command flag command interface
type Command interface {
	register()           // implement it to read flag values
	listen() (exit bool) // handle cmdName, return true when program should exit
}

var commands = []Command{
	&versionCmd{},
	&createCmd{},
}

// Listen listen commands input
func Listen() bool {
	// Read cmdName via flag
	for i := 0; i < len(commands); i++ {
		commands[i].register()
	}
	flag.Parse()

	// Receive commands
	for i := 0; i < len(commands); i++ {
		if commands[i].listen() {
			return true
		}
	}

	return false
}
