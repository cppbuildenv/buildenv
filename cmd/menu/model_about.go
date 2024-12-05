package menu

import (
	"buildenv/config"

	tea "github.com/charmbracelet/bubbletea"
)

func newAboutModel(callbacks config.PlatformCallbacks, goback func()) *aboutModel {
	return &aboutModel{
		callbacks: callbacks,
		goback:    goback,
	}
}

type aboutModel struct {
	callbacks config.PlatformCallbacks
	goback    func()
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
			u.goback()
			return u, nil
		}
	}
	return u, nil
}

func (u aboutModel) View() string {
	return u.callbacks.About()
}
