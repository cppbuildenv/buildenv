package menu

import (
	"buildenv/config"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	currentMode mode
)

func CreateMainModel(callabcks config.BuildEnvCallbacks) MainModel {
	return MainModel{
		menuMode: createMenuModel(func(mode mode) {
			currentMode = mode
		}),
		syncModel: newSyncModel(func() {
			currentMode = modeMenu
		}),
		platformCreateModel: newPlatformCreateModel(callabcks, func(this *platformCreateModel) {
			this.Reset()
			currentMode = modeMenu
		}),
		platformSelectModel: newPlatformSelectModel(callabcks, func() {
			currentMode = modeMenu
		}),
		projectCreateModel: newProjectCreateModel(callabcks, func(this *projectCreateModel) {
			this.Reset()
			currentMode = modeMenu
		}),
		projectSelectModel: newProjectSelectModel(callabcks, func() {
			currentMode = modeMenu
		}),
		integrateModel: newIntegrateModel(func() {
			currentMode = modeMenu
		}),
		aboutModel: newAboutModel(callabcks, func() {
			currentMode = modeMenu
		}),
	}
}

type MainModel struct {
	menuMode            tea.Model
	syncModel           tea.Model
	platformCreateModel tea.Model
	platformSelectModel tea.Model
	projectCreateModel  tea.Model
	projectSelectModel  tea.Model
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
			model, cmd := m.syncModel.Update(msg)
			m.syncModel = model
			return m, cmd

		case modePlatformCreate:
			model, cmd := m.platformCreateModel.Update(msg)
			m.platformCreateModel = model
			return m, cmd

		case modePlatformSelect:
			model, cmd := m.platformSelectModel.Update(msg)
			m.platformSelectModel = model
			return m, cmd

		case modeProjectCreate:
			model, cmd := m.projectCreateModel.Update(msg)
			m.projectCreateModel = model
			return m, cmd

		case modeProjectSelect:
			model, cmd := m.projectSelectModel.Update(msg)
			m.projectSelectModel = model
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
		return m.syncModel.View()

	case modePlatformCreate:
		return m.platformCreateModel.View()

	case modePlatformSelect:
		return m.platformSelectModel.View()

	case modeProjectCreate:
		return m.projectCreateModel.View()

	case modeProjectSelect:
		return m.projectSelectModel.View()

	case modelIntegrate:
		return m.integrateModel.View()

	case modeAbout:
		return m.aboutModel.View()
	}

	return ""
}
