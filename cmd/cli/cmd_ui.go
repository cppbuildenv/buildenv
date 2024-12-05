package cli

import (
	uipkg "buildenv/cmd/menu"
	"buildenv/config"
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

func (u *uiCmd) register() {
	flag.BoolVar(&u.interactive, "ui", false, "run buildenv in gui mode.")
}

func (u *uiCmd) listen() (handled bool) {
	if !u.interactive {
		return false
	}

	model := uipkg.CreateMainModel(u.platformCallbacks)
	if _, err := tea.NewProgram(model).Run(); err != nil {
		log.Fatalf("Running cli in gui mode error: %s", err)
	}
	return true
}
