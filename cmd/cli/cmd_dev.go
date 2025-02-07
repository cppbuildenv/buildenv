package cli

import "flag"

func newDevCmd() *devCmd {
	return &devCmd{}
}

type devCmd struct {
	dev bool
}

func (b *devCmd) register() {
	flag.BoolVar(&b.dev, "dev", false, "install or remove a third-party as dev, works with -'-install', '--remove' and '--purge'.")
}
