package menu

import (
	"buildenv/config"
	"slices"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

var MenuModel = newMenuModel(config.Callbacks)

const (
	menuInit           string = "Init buildenv with conf repo."
	menuPlatformCreate string = "Create a new platform."
	menuPlatformSelect string = "Select your current platform."
	menuProjectCreate  string = "Create a new project."
	menuProjectSelect  string = "Select your current project."
	menuToolCreate     string = "Create a new tool."
	menuPortCreate     string = "Create a new port."
	menuIntegrate      string = "Integrate buildenv, then you can run it everywhere."
	menuAbout          string = "About and usage."
)

var menus = []string{
	menuInit,
	menuPlatformCreate,
	menuPlatformSelect,
	menuProjectCreate,
	menuProjectSelect,
	menuToolCreate,
	menuPortCreate,
	menuIntegrate,
	menuAbout,
}

type errMsg error

func newMenuModel(callabcks config.BuildEnvCallbacks) *menuModel {
	const defaultWidth = 100
	const defaultHeight = 18

	var items []list.Item
	for _, menu := range menus {
		items = append(items, listItem(menu))
	}

	l := list.New(items, listDelegate{}, defaultWidth, defaultHeight)
	l.Title = "Welcome to buildenv! \nPlease choose an option from the menu below..."

	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	menuModel := menuModel{
		list:   l,
		models: make(map[string]tea.Model),
	}

	// init models
	menuModel.models[menuInit] = newInitModel(callabcks)
	menuModel.models[menuPlatformCreate] = newPlatformCreateModel(callabcks)
	menuModel.models[menuPlatformSelect] = newPlatformSelectModel(callabcks)
	menuModel.models[menuProjectCreate] = newProjectCreateModel(callabcks)
	menuModel.models[menuProjectSelect] = newProjectSelectModel(callabcks)
	menuModel.models[menuToolCreate] = newToolCreateModel(callabcks)
	menuModel.models[menuPortCreate] = newPortCreateModel(callabcks)
	menuModel.models[menuIntegrate] = newIntegrateModel()
	menuModel.models[menuAbout] = newAboutModel(callabcks)

	return &menuModel
}

type menuModel struct {
	list   list.Model
	models map[string]tea.Model
}

func (m menuModel) Init() tea.Cmd {
	return nil
}

func (m *menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return m, tea.Quit

		case "enter":
			if i, ok := m.list.SelectedItem().(listItem); ok {
				// Remember selected item.
				index := slices.Index(menus, string(i))
				m.list.Select(index)

				return m.models[string(i)], nil
			}

			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m menuModel) View() string {
	return "\n" + m.list.View()
}
