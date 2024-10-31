package interactive

import (
	tea "github.com/charmbracelet/bubbletea"
)

var (
	currentMode mode
)

func CreateMainModel(callabcks CommandCallbacks) MainModel {
	goback := func() { currentMode = modeMenu }

	return MainModel{
		optionModel: createMenuModel(func(mode mode) {
			currentMode = mode
		}),
		platformEditModel: createPlatformCreateModel(func(name string) error {
			return callabcks.OnCreatePlatform(name)
		}, goback),
		platformPickModel: createPlatformPickModel("./conf/platform", func(platform string) {
			callabcks.OnPickPlatform(platform)
		}, goback),
		aboutModel: createAboutModel(goback),
	}
}

type CommandCallbacks interface {
	OnCreatePlatform(platformName string) error
	OnPickPlatform(platformName string)
}

type MainModel struct {
	optionModel       tea.Model
	platformEditModel tea.Model
	platformPickModel tea.Model
	aboutModel        tea.Model
}

func (m MainModel) Init() tea.Cmd {
	return nil
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch currentMode {
		case modeMenu:
			updatedModel, cmd := m.optionModel.Update(msg)
			m.optionModel = updatedModel
			return m, cmd

		case modePlatformEdit:
			updatedModel, cmd := m.platformEditModel.Update(msg)
			m.platformEditModel = updatedModel
			return m, cmd

		case modePlatformList:
			updatedModel, cmd := m.platformPickModel.Update(msg)
			m.platformPickModel = updatedModel
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
		return m.optionModel.View()

	case modePlatformEdit:
		return m.platformEditModel.View()

	case modePlatformList:
		return m.platformPickModel.View()

	case modeAbout:
		return m.aboutModel.View()
	}

	return ""
}
