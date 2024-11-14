package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	menuSyncConfig     string = "Init or sync config."
	menuChoosePlatform string = "Choose a platform as your build target platform."
	menuCreatePlatform string = "Create a new platform, it requires completion later."
	menuUsage          string = "About and Usage."
)

type mode = int

const (
	modeMenu mode = iota
	modeSyncConfig
	modePlatformChoose
	modePlatformCreate
	modeAbout
)

var menus = []string{
	menuSyncConfig,
	menuChoosePlatform,
	menuCreatePlatform,
	menuUsage,
}

func createMenuModel(modeChanged func(mode mode)) menuModel {
	const defaultWidth = 80
	const defaultHeight = 10

	var items []list.Item
	for _, menu := range menus {
		items = append(items, listItem(menu))
	}

	styles := createStyles()

	l := list.New(items, listDelegate{styles}, defaultWidth, defaultHeight)
	l.Title = "Please choose one from the menu..."

	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	l.Styles.Title = styles.titleStyle
	l.Styles.PaginationStyle = styles.paginationStyle
	l.Styles.HelpStyle = styles.helpStyle

	return menuModel{list: l, modeChanged: modeChanged, styles: styles}
}

type menuModel struct {
	list        list.Model
	quitting    bool
	styles      styles
	modeChanged func(mode mode)
}

func (m menuModel) Init() tea.Cmd {
	return nil
}

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if i, ok := m.list.SelectedItem().(listItem); ok {
				if m.modeChanged != nil {
					switch string(i) {
					case menuSyncConfig:
						m.modeChanged(modeSyncConfig)

					case menuChoosePlatform:
						m.modeChanged(modePlatformChoose)

					case menuCreatePlatform:
						m.modeChanged(modePlatformCreate)

					case menuUsage:
						m.modeChanged(modeAbout)
					}
				}
			}

			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m menuModel) View() string {
	if m.quitting {
		return ""
	}

	return "\n" + m.list.View()
}
