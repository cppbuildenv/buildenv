package menu

import (
	"buildenv/config"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func newProjectCreateModel(callbacks config.BuildEnvCallbacks) *projectCreateModel {
	textInput := textinput.New()
	textInput.Placeholder = "your project's name..."
	textInput.Focus()
	textInput.CharLimit = 100
	textInput.Width = 100
	textInput.TextStyle = focusedStyle
	textInput.PromptStyle = focusedStyle
	textInput.Cursor.Style = focusedStyle

	return &projectCreateModel{
		textInput: textInput,
		finished:  false,
		err:       nil,
		callbacks: callbacks,
	}
}

type projectCreateModel struct {
	textInput textinput.Model
	finished  bool
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
			p.reset()
			return MenuModel, nil

		case "enter":
			// Clean platform name.
			projectName := strings.TrimSpace(p.textInput.Value())
			projectName = strings.TrimSuffix(projectName, ".json")
			p.textInput.SetValue(projectName)

			if err := p.callbacks.OnCreateProject(p.textInput.Value()); err != nil {
				p.err = err
				p.finished = false
			} else {
				p.finished = true
			}

			return p, tea.Quit
		}
	}

	var cmd tea.Cmd
	p.textInput, cmd = p.textInput.Update(msg)
	return p, cmd
}

func (p projectCreateModel) View() string {
	if p.finished {
		return config.SprintSuccess("%s is created but need to config it later.", p.textInput.Value())
	}

	if p.err != nil {
		return config.SprintError(p.err, "%s could not be created.", p.textInput.Value())
	}

	return fmt.Sprintf("\n%s\n\n%s\n\n%s\n",
		titleStyle.Width(50).Render("Please enter your project's name:"),
		p.textInput.View(),
		actionBarStyle.Render("[enter -> execute | esc -> back | ctrl+c/q -> quit]"),
	)
}

func (p *projectCreateModel) reset() {
	p.textInput.Reset()
	p.finished = false
	p.err = nil
}
