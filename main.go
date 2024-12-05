package main

import (
	"buildenv/cmd/cli"
	"buildenv/cmd/menu"
	"buildenv/config"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if handled := cli.Listen(); handled {
		os.Exit(0)
	}

	// Run in ui mode in default.
	model := menu.CreateMainModel(config.Callbacks)
	if _, err := tea.NewProgram(model).Run(); err != nil {
		log.Fatalf("run buildenv in ui mode: %s", err)
	}
}
