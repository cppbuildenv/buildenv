package cli

import "flag"

func newSilentModeCmd() *silentModeCmd {
	return &silentModeCmd{}
}

type silentModeCmd struct {
	silent bool
}

func (s *silentModeCmd) register() {
	flag.BoolVar(&s.silent, "silent", false, "run cli in silent mode")
}
