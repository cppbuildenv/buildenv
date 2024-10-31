package cli

import "flag"

func newForceCmd() *forceCmd {
	return &forceCmd{}
}

type forceCmd struct {
	force bool
}

func (cmd *forceCmd) register() {
	flag.BoolVar(&cmd.force, "f", false, "execute command forcely")
	flag.BoolVar(&cmd.force, "force", false, "execute command forcely")
}
