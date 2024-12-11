package main

import (
	"buildenv/cmd/cli"
	"buildenv/cmd/menu"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if handled := cli.Listen(); handled {
		os.Exit(0)
	}

	// Run in ui mode in default.
	if _, err := tea.NewProgram(menu.MenuModel).Run(); err != nil {
		log.Fatalf("run buildenv in ui mode: %s", err)
	}
}
