package main

import (
	"buildenv/cmd/cli"
	"buildenv/cmd/menu"
	"log"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("get home directory: %s", err)
	}

	// Clean environment.
	os.Clearenv()

	var paths []string
	paths = append(paths, "/usr/local/bin")
	paths = append(paths, "/usr/bin")
	paths = append(paths, "/usr/sbin")
	paths = append(paths, homeDir+"/.local/bin")
	os.Setenv("PATH", strings.Join(paths, string(os.PathListSeparator)))

	// Listen for cli request.
	if handled := cli.Listen(); handled {
		os.Exit(0)
	}

	// Run in ui mode in default.
	if _, err := tea.NewProgram(menu.MenuModel).Run(); err != nil {
		log.Fatalf("run buildenv in ui mode: %s", err)
	}
}
