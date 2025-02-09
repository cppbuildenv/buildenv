package main

import (
	"buildenv/cmd/cli"
	"buildenv/cmd/menu"
	"buildenv/config"
	"fmt"
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

	if len(os.Args) == 1 {
		// Run in ui mode in default.
		if _, err := tea.NewProgram(menu.MenuModel).Run(); err != nil {
			log.Fatalf("run buildenv in menu mode: %s", err)
		}
	} else if os.Args[1] == "--help" || os.Args[1] == "-h" {
		printUsage()
	} else {
		cmdName := os.Args[1]
		for _, cmd := range cli.Commands {
			if cmd.Name == cmdName {
				cmd.Handler(config.Callbacks)
				return
			}
		}

		fmt.Printf("Unknown command: %s\n\n", cmdName)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf("Usage: %s [command] [options]\n\n", os.Args[0])
	fmt.Println("Available commands:")
	for _, cmd := range cli.Commands {
		fmt.Printf("  %-10s%s\n", cmd.Name, cmd.Description)
	}
	fmt.Println("\nRun './buildenv [command] --help' for more information about a command.")
}
