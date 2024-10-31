package cli

import (
	inter "buildenv/console/interactive"
	"flag"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

func newInteractiveCmd() *interactiveCmd {
	return &interactiveCmd{}
}

type interactiveCmd struct {
	interactive bool
}

func (cmd *interactiveCmd) register() {
	flag.BoolVar(&cmd.interactive, "i", false, "run in interactive mode")
	flag.BoolVar(&cmd.interactive, "interactive", false, "run in interactive mode")
}

func (cmd *interactiveCmd) listen() (handled bool) {
	if !cmd.interactive {
		return false
	}

	model := inter.CreateMainModel(commondCallbacks{})
	if _, err := tea.NewProgram(model).Run(); err != nil {
		log.Fatalf("Running cli in interactive mode error: %s", err)
	}
	return true
}

type commondCallbacks struct {
}

func (c commondCallbacks) OnCreatePlatform(platformName string) error {
	if platformName == "" {
		return fmt.Errorf("platform name is empty")
	}

	return nil
}

func (c commondCallbacks) OnSelectPlatform(platformName string) {
}
