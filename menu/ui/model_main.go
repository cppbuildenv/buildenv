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
		syncConfigModel: newSyncConfigModel(func() {
			currentMode = modeMenu
		}),
		platformCreateModel: newPlatformCreateModel(callabcks, func(this *platformCreateModel) {
			this.Reset()
			currentMode = modeMenu
		}),
		platformSelectModel: newPlatformSelectModel(callabcks, func() {
			currentMode = modeMenu
		}),
		integrateModel: newIntegrateModel(func() {
			currentMode = modeMenu
		}),
		aboutModel: newUsageModel(func() {
			currentMode = modeMenu
		}),
	}
}

type MainModel struct {
	menuMode            tea.Model
	syncConfigModel     tea.Model
	platformCreateModel tea.Model
	platformSelectModel tea.Model
	integrateModel      tea.Model
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
			model, cmd := m.menuMode.Update(msg)
			m.menuMode = model
			return m, cmd

		case modeSyncConfig:
			model, cmd := m.syncConfigModel.Update(msg)
			m.syncConfigModel = model
			return m, cmd

		case modePlatformCreate:
			model, cmd := m.platformCreateModel.Update(msg)
			m.platformCreateModel = model
			return m, cmd

		case modePlatformSelect:
			model, cmd := m.platformSelectModel.Update(msg)
			m.platformSelectModel = model
			return m, cmd

		case modelIntegrate:
			model, cmd := m.integrateModel.Update(msg)
			m.integrateModel = model
			return m, cmd

		case modeAbout:
			model, cmd := m.aboutModel.Update(msg)
			m.aboutModel = model
			return m, cmd
		}
	}

	return m, nil
}

func (m MainModel) View() string {
	switch currentMode {
	case modeMenu:
		return m.menuMode.View()

	case modeSyncConfig:
		return m.syncConfigModel.View()

	case modePlatformCreate:
		return m.platformCreateModel.View()

	case modePlatformSelect:
		return m.platformSelectModel.View()

	case modelIntegrate:
		return m.integrateModel.View()

	case modeAbout:
		return m.aboutModel.View()
	}

	return ""
}
