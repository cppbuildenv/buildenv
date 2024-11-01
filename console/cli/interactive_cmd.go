package cli

import (
	"buildenv/config"
	inter "buildenv/console/interactive"
	"flag"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

func newInteractiveCmd(platformCallbacks config.PlatformCallbacks) *interactiveCmd {
	return &interactiveCmd{
		platformCallbacks: platformCallbacks,
	}
}

type interactiveCmd struct {
	interactive       bool
	platformCallbacks config.PlatformCallbacks
}

func (cmd *interactiveCmd) register() {
	flag.BoolVar(&cmd.interactive, "i", false, "run in interactive mode")
	flag.BoolVar(&cmd.interactive, "interactive", false, "run in interactive mode")
}

func (cmd *interactiveCmd) listen() (handled bool) {
	if !cmd.interactive {
		return false
	}

	model := inter.CreateMainModel(cmd.platformCallbacks)
	if _, err := tea.NewProgram(model).Run(); err != nil {
		log.Fatalf("Running cli in interactive mode error: %s", err)
	}
	return true
}
