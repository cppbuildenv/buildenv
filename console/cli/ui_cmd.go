package cli

import (
	"buildenv/config"
	"buildenv/console/ui"
	"flag"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

func newUICmd(platformCallbacks config.PlatformCallbacks) *uiCmd {
	return &uiCmd{
		platformCallbacks: platformCallbacks,
	}
}

type uiCmd struct {
	interactive       bool
	platformCallbacks config.PlatformCallbacks
}

func (cmd *uiCmd) register() {
	flag.BoolVar(&cmd.interactive, "ui", false, "run in ui mode")
}

func (cmd *uiCmd) listen() (handled bool) {
	if !cmd.interactive {
		return false
	}

	model := ui.CreateMainModel(cmd.platformCallbacks)
	if _, err := tea.NewProgram(model).Run(); err != nil {
		log.Fatalf("Running cli in interactive mode error: %s", err)
	}
	return true
}
