package menu

import (
	"buildenv/config"
	"buildenv/pkg/fileio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func newProjectSelectModel(callbacks config.BuildEnvCallbacks) *projectSelectModel {
	const defaultWidth = 80
	const defaultHeight = 10

	var items []list.Item

	if fileio.PathExists(config.Dirs.ProjectsDir) {
		// List all entities in project dir.
		entities, err := os.ReadDir(config.Dirs.ProjectsDir)
		if err != nil {
			fmt.Println("Error reading projects dir:", err)
			os.Exit(1)
		}

		// Create list items with name of entities.
		for _, entity := range entities {
			if !entity.IsDir() && strings.HasSuffix(entity.Name(), ".json") {
				projectName := strings.TrimSuffix(entity.Name(), ".json")
				items = append(items, listItem(projectName))
			}
		}
	}

	l := list.New(items, listDelegate{}, defaultWidth, defaultHeight)
	l.Title = "Select your current project:"

	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return &projectSelectModel{
		list:      l,
		callbacks: callbacks,
	}
}

type projectSelectModel struct {
	list        list.Model
	trySelected string
	selected    string
	err         error
	callbacks   config.BuildEnvCallbacks
}

func (p projectSelectModel) Init() tea.Cmd {
	return nil
}

func (p projectSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.list.SetWidth(msg.Width)
		return p, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if i, ok := p.list.SelectedItem().(listItem); ok {
				p.trySelected = string(i)
				if err := p.callbacks.OnSelectProject(string(i)); err != nil {
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

func (p projectSelectModel) View() string {
	if p.err != nil {
		return config.SprintError(p.err, "failed to select project: %s.", p.trySelected)
	}

	if p.selected != "" {
		return config.SprintSuccess("buildenv is ready for project: %s.", p.selected)
	}

	return "\n" + p.list.View()
}
