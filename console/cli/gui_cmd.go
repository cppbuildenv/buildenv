package cli

import (
	"buildenv/config"
	"buildenv/console/ui"
	"flag"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

func newGUICmd(platformCallbacks config.PlatformCallbacks) *guiCmd {
	return &guiCmd{
		platformCallbacks: platformCallbacks,
	}
}

type guiCmd struct {
	interactive       bool
	platformCallbacks config.PlatformCallbacks
}

func (g *guiCmd) register() {
	flag.BoolVar(&g.interactive, "gui", false, "run in gui mode")
}

func (g *guiCmd) listen() (handled bool) {
	if !g.interactive {
		return false
	}

	model := ui.CreateMainModel(g.platformCallbacks)
	if _, err := tea.NewProgram(model).Run(); err != nil {
		log.Fatalf("Running cli in gui mode error: %s", err)
	}
	return true
}
