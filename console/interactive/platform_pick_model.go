package interactive

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func createPlatformPickModel(platformDir string, picked func(platform string), goback func()) platformPickModel {
	const defaultWidth = 80
	const defaultHeight = 10

	// Create platform dir if not exists
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
	l.Title = "Please pick one platform as your build target platform:"

	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	l.Styles.Title = styleImpl.titleStyle
	l.Styles.PaginationStyle = styleImpl.paginationStyle
	l.Styles.HelpStyle = styleImpl.helpStyle

	return platformPickModel{
		list:   l,
		styles: styleImpl,
		picked: picked,
		goback: goback,
	}
}

type platformPickModel struct {
	list   list.Model
	value  string
	styles styles
	picked func(platform string)
	goback func()
}

func (p platformPickModel) Init() tea.Cmd {
	return nil
}

func (p platformPickModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.list.SetWidth(msg.Width)
		return p, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if i, ok := p.list.SelectedItem().(listItem); ok {
				p.value = string(i)
				p.picked(p.value)
			}
			return p, tea.Quit

		case "ctrl+c", "esc", "q":
			p.goback()
			return p, nil
		}
	}

	var cmd tea.Cmd
	p.list, cmd = p.list.Update(msg)
	return p, cmd
}

func (p platformPickModel) View() string {
	if p.value != "" {
		return p.styles.quitTextStyle.Render(fmt.Sprintf("[✔] ---- build target platform: %s", p.value))
	}

	return "\n" + p.list.View()
}
