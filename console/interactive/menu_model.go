package interactive

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	menuCreatePlatform string = "Create a new platform."
	menuChoosePlatform string = "Choose a platform as your build target."
	menuAbout          string = "About buildenv."
)

type mode = int

const (
	modeMenu mode = iota
	modePlatformEdit
	modePlatformList
	modeAbout
)

var options = []string{
	menuCreatePlatform,
	menuChoosePlatform,
	menuAbout,
}

func createMenuModel(modeChanged func(mode mode)) menuModel {
	const defaultWidth = 80
	const defaultHeight = 10

	var items []list.Item
	for _, option := range options {
		items = append(items, listItem(option))
	}

	styles := createStyles()

	l := list.New(items, listDelegate{styles}, defaultWidth, defaultHeight)
	l.Title = "Please choose one option from the list..."

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
	value       string
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
				m.value = string(i)
				if m.modeChanged != nil {
					switch m.value {
					case menuCreatePlatform:
						m.modeChanged(modePlatformEdit)

					case menuChoosePlatform:
						m.modeChanged(modePlatformList)

					case menuAbout:
						m.modeChanged(modeAbout)
					}
				}

				m.value = ""
			}

			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m menuModel) View() string {
	if m.value != "" {
		return m.styles.quitTextStyle.Render(fmt.Sprintf("You selected: %s", m.value))
	}

	if m.quitting {
		return m.styles.quitTextStyle.Render("See you next time...")
	}

	return "\n" + m.list.View()
}
