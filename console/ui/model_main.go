package ui

import (
	"buildenv/config"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	currentMode mode
)

func CreateMainModel(callabcks config.PlatformCallbacks) MainModel {
	return MainModel{
		menuMode: createMenuModel(func(mode mode) {
			currentMode = mode
		}),
		platformCreateModel: newPlatformCreateModel(callabcks, func(this *platformCreateModel) {
			this.Reset()
			currentMode = modeMenu
		}),
		platformSelectModel: newPlatformSelectModel(callabcks, func() {
			currentMode = modeMenu
		}),
		aboutModel: newUsageModel(func() {
			currentMode = modeMenu
		}),
	}
}

type MainModel struct {
	menuMode            tea.Model
	platformCreateModel tea.Model
	platformSelectModel tea.Model
	aboutModel          tea.Model
}

func (m MainModel) Init() tea.Cmd {
	return nil
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch currentMode {
		case modeMenu:
			updatedModel, cmd := m.menuMode.Update(msg)
			m.menuMode = updatedModel
			return m, cmd

		case modePlatformCreate:
			updatedModel, cmd := m.platformCreateModel.Update(msg)
			m.platformCreateModel = updatedModel
			return m, cmd

		case modePlatformChoose:
			updatedModel, cmd := m.platformSelectModel.Update(msg)
			m.platformSelectModel = updatedModel
			return m, cmd

		case modeAbout:
			updatedModel, cmd := m.aboutModel.Update(msg)
			m.aboutModel = updatedModel
			return m, cmd
		}
	}

	return m, nil
}

func (m MainModel) View() string {
	switch currentMode {
	case modeMenu:
		return m.menuMode.View()

	case modePlatformCreate:
		return m.platformCreateModel.View()

	case modePlatformChoose:
		return m.platformSelectModel.View()

	case modeAbout:
		return m.aboutModel.View()
	}

	return ""
}
