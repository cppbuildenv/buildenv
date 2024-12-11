package menu

import (
	"buildenv/config"
	"buildenv/pkg/color"
	"buildenv/pkg/env"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

func newIntegrateModel() integrateModel {
	content := fmt.Sprintf("\nIntegrate buildenv.\n"+
		"-----------------------------------\n"+
		"%s.\n\n"+
		"%s",
		color.Sprintf(color.Blue, "This will append buildenv's file dir to ~/.profile, then you can use buildenv anywhere."),
		color.Sprintf(color.Gray, "[↵ -> execute | ctrl+c/q -> quit]"))
	return integrateModel{content: content}
}

type integrateModel struct {
	content string
}

func (i integrateModel) Init() tea.Cmd {
	return nil
}

func (i integrateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return i, tea.Quit

		case "enter":
			i.integrate()
			return i, tea.Quit

		case "esc":
			return MenuModel, nil
		}
	}
	return i, nil
}

func (i integrateModel) View() string {
	return i.content
}

func (i integrateModel) integrate() {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Print(config.IntegrateFailed(err))
		os.Exit(1)
	}

	if err := env.UpdateRunPath(filepath.Dir(exePath)); err != nil {
		fmt.Print(config.IntegrateFailed(err))
		os.Exit(1)
	}

	fmt.Print(config.IntegrateSuccess())
}
