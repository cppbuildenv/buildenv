package menu

import (
	"buildenv/config"
	"buildenv/pkg/color"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func newProjectCreateModel(callbacks config.BuildEnvCallbacks) *projectCreateModel {
	ti := textinput.New()
	ti.Placeholder = "your project's name..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 100
	ti.TextStyle = styleImpl.focusedStyle
	ti.PromptStyle = styleImpl.focusedStyle
	ti.Cursor.Style = styleImpl.focusedStyle

	return &projectCreateModel{
		textInput: ti,
		callbacks: callbacks,
	}
}

type projectCreateModel struct {
	textInput textinput.Model
	created   bool
	err       error

	callbacks config.BuildEnvCallbacks
}

func (p projectCreateModel) Init() tea.Cmd {
	return textinput.Blink
}

func (p projectCreateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return p, tea.Quit

		case "esc":
			return MenuModel, nil

		case "enter":
			// Clean platform name.
			projectName := strings.TrimSpace(p.textInput.Value())
			projectName = strings.TrimSuffix(projectName, ".json")
			p.textInput.SetValue(projectName)

			if err := p.callbacks.OnCreateProject(p.textInput.Value()); err != nil {
				p.err = err
				p.created = false
			} else {
				p.created = true
			}

			return p, tea.Quit
		}
	}

	var cmd tea.Cmd
	p.textInput, cmd = p.textInput.Update(msg)
	return p, cmd
}

func (p projectCreateModel) View() string {
	if p.created {
		return config.ProjectCreated(p.textInput.Value())
	}

	if p.err != nil {
		return config.ProjectCreateFailed(p.textInput.Value(), p.err)
	}

	return fmt.Sprintf("\n%s\n\n%s\n\n%s\n",
		color.Sprintf(color.Blue, "Please enter your project's name: "),
		p.textInput.View(),
		color.Sprintf(color.Gray, "[esc -> back | ctrl+c/q -> quit]"),
	)
}

func (p *projectCreateModel) Reset() {
	p.textInput.Reset()
	p.created = false
	p.err = nil
}
