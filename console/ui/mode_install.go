package ui

import (
	"buildenv/console"
	"buildenv/pkg/color"
	"buildenv/pkg/env"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

func newInstallModel(goback func()) installModel {
	content := fmt.Sprintf("\nInstall buildenv.\n"+
		"-----------------------------------\n"+
		"%s.\n\n"+
		"%s",
		color.Sprintf(color.Blue, "This will add buildenv's path to your PATH, so that you can use buildenv anywhere."),
		color.Sprintf(color.Gray, "[↵ -> execute | ctrl+c/q -> quit]"))
	return installModel{
		content: content,
		goback:  goback,
	}
}

type installModel struct {
	content string
	goback  func()
}

func (i installModel) Init() tea.Cmd {
	return nil
}

func (i installModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return i, tea.Quit

		case "enter":
			i.install()
			return i, tea.Quit

		case "esc":
			i.goback()
			return i, nil
		}
	}
	return i, nil
}

func (i installModel) View() string {
	return i.content
}

func (i installModel) install() {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Print(console.InstallFailed(err))
		os.Exit(1)
	}

	if err := env.UpdateRunPath(filepath.Dir(exePath)); err != nil {
		fmt.Print(console.InstallFailed(err))
		os.Exit(1)
	}

	fmt.Print(console.InstallSuccess())
}
