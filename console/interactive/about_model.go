package interactive

import (
	tea "github.com/charmbracelet/bubbletea"
)

func createAboutModel(goback func()) aboutModel {
	content := `
Welcome to the buildenv.
-----------------------------------

This is a simple tool to manage your cross build environment.

[press ctrl+c or q to quit]`
	return aboutModel{content: content, goback: goback}
}

type aboutModel struct {
	content string
	goback  func()
}

func (aboutModel) Init() tea.Cmd {
	return nil
}

func (a aboutModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit

		case "esc":
			a.goback()
			return a, nil
		}
	}
	return a, nil
}

func (a aboutModel) View() string {
	return a.content
}
