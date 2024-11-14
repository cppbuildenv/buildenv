package cli

import "flag"

func newSilentCmd() *silentCmd {
	return &silentCmd{}
}

type silentCmd struct {
	silent bool
}

func (s *silentCmd) register() {
	flag.BoolVar(&s.silent, "silent", false, "run in silent mode")
}
