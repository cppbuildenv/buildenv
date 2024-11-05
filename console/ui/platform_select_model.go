package ui

import (
	"buildenv/config"
	"buildenv/console"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func createPlatformSelectModel(platformDir string, callbacks config.PlatformCallbacks, goback func()) platformSelectModel {
	const defaultWidth = 80
	const defaultHeight = 10

	// Create platform dir if not exists.
	if err := os.MkdirAll(platformDir, 0755); err != nil {
		fmt.Println("Error creating platform dir:", err)
		os.Exit(1)
	}

	// List all entities in platform dir.
	entities, err := os.ReadDir(platformDir)
	if err != nil {
		fmt.Println("Error reading platform dir:", err)
		os.Exit(1)
	}

	// Create list items with name of entities.
	var items []list.Item
	for _, entity := range entities {
		if !entity.IsDir() && strings.HasSuffix(entity.Name(), ".json") {
			simpleName := strings.TrimSuffix(entity.Name(), ".json")
			items = append(items, listItem(simpleName))
		}
	}

	l := list.New(items, listDelegate{styleImpl}, defaultWidth, defaultHeight)
	l.Title = "Please select one platform as your build target platform:"

	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	l.Styles.Title = styleImpl.titleStyle
	l.Styles.PaginationStyle = styleImpl.paginationStyle
	l.Styles.HelpStyle = styleImpl.helpStyle

	return platformSelectModel{
		list:      l,
		styles:    styleImpl,
		callbacks: callbacks,
		goback:    goback,
	}
}

type platformSelectModel struct {
	list      list.Model
	value     string
	err       error
	styles    styles
	callbacks config.PlatformCallbacks
	goback    func()
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
				filePath := filepath.Join(config.PlatformsDir, string(i)+".json")
				if err := p.callbacks.OnSelectPlatform(filePath); err != nil {
					p.err = err
				} else {
					p.value = string(i)
					p.err = nil
				}
			}
			return p, tea.Quit

		case "ctrl+c", "esc", "q":
			p.goback()
			p.value = ""
			p.err = nil
			return p, nil
		}
	}

	var cmd tea.Cmd
	p.list, cmd = p.list.Update(msg)
	return p, cmd
}

func (p platformSelectModel) View() string {
	if p.err != nil {
		return p.styles.resultTextStyle.Render(fmt.Sprintf(console.PlatformSelectedFailed, p.value, p.err))
	}

	if p.value != "" {
		return p.styles.resultTextStyle.Render(fmt.Sprintf(console.PlatformSelected, p.value))
	}

	return "\n" + p.list.View()
}
