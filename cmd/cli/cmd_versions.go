package cli

import (
	"flag"
	"fmt"
)

func newVersionCmd() *versionCmd {
	return &versionCmd{}
}

var (
	Version string // for example: `1.0.0`
)

type versionCmd struct {
	version bool
}

func (v *versionCmd) register() {
	flag.BoolVar(&v.version, "version", false, "print version.")
}

func (cmd *versionCmd) listen() (handled bool) {
	if cmd.version {
		cmd.PrintVersion()
		return true
	}

	return false
}

// PrintVersion print versions when app launch
func (v *versionCmd) PrintVersion() {
	if Version != "" {
		fmt.Printf("Version: %s\n", Version)
	}
}
