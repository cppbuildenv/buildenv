package cli

import (
	"buildenv/config"
	uipkg "buildenv/console/ui"
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
	flag.BoolVar(&u.interactive, "gui", false, "run in gui mode")
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
