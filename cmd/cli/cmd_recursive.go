package cli

import "flag"

func newRecursiveCmd() *recursiveCmd {
	return &recursiveCmd{}
}

type recursiveCmd struct {
	recursive bool
}

func (r *recursiveCmd) register() {
	flag.BoolVar(&r.recursive, "recursive", false, "it works with -uninstall.")
}
