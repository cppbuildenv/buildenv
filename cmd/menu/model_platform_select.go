package menu

import (
	"buildenv/config"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func newPlatformSelectModel(callbacks config.BuildEnvCallbacks) *platformSelectModel {
	const defaultWidth = 80
	const defaultHeight = 10

	// Create platform dir if not exists.
	if err := os.MkdirAll(config.Dirs.PlatformsDir, 0755); err != nil {
		fmt.Println("Error creating platform dir:", err)
		os.Exit(1)
	}

	// List all entities in platform dir.
	entities, err := os.ReadDir(config.Dirs.PlatformsDir)
	if err != nil {
		fmt.Println("Error reading platform dir:", err)
		os.Exit(1)
	}

	// Create list items with name of entities.
	var items []list.Item
	for _, entity := range entities {
		if !entity.IsDir() && strings.HasSuffix(entity.Name(), ".json") {
			platformName := strings.TrimSuffix(entity.Name(), ".json")
			items = append(items, listItem(platformName))
		}
	}

	l := list.New(items, listDelegate{}, defaultWidth, defaultHeight)
	l.Title = "Select your current platform:"

	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return &platformSelectModel{
		list:      l,
		callbacks: callbacks,
	}
}

type platformSelectModel struct {
	list        list.Model
	trySelected string
	selected    string
	err         error
	callbacks   config.BuildEnvCallbacks
}

func (p platformSelectModel) Init() tea.Cmd {
	return nil
}

func (p platformSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.list.SetWidth(msg.Width)
		return p, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if i, ok := p.list.SelectedItem().(listItem); ok {
				p.trySelected = string(i)
				if err := p.callbacks.OnSelectPlatform(string(i)); err != nil {
					p.err = err
				} else {
					p.selected = string(i)
					p.err = nil
				}
			}
			return p, tea.Quit

		case "esc":
			p.trySelected = ""
			p.selected = ""
			p.err = nil
			return MenuModel, nil

		case "ctrl+c", "q":
			return p, tea.Quit
		}
	}

	var cmd tea.Cmd
	p.list, cmd = p.list.Update(msg)
	return p, cmd
}

func (p platformSelectModel) View() string {
	if p.err != nil {
		return config.PlatformSelectedFailed(p.trySelected, p.err)
	}

	if p.selected != "" {
		return config.PlatformSelected(p.selected)
	}

	return "\n" + p.list.View()
}
