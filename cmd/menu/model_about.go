package menu

import (
	"buildenv/cmd/cli"
	"buildenv/config"

	tea "github.com/charmbracelet/bubbletea"
)

func newAboutModel(callbacks config.BuildEnvCallbacks) *aboutModel {
	return &aboutModel{
		callbacks: callbacks,
	}
}

type aboutModel struct {
	callbacks config.BuildEnvCallbacks
}

func (aboutModel) Init() tea.Cmd {
	return nil
}

func (u aboutModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return u, tea.Quit

		case "esc":
			return MenuModel, nil
		}
	}
	return u, nil
}

func (u aboutModel) View() string {
	return u.callbacks.About(cli.Version)
}
