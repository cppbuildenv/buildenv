package menu

import (
	"buildenv/config"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func newPlatformCreateModel(callbacks config.BuildEnvCallbacks) *platformCreateModel {
	textInput := textinput.New()
	textInput.Placeholder = "for example: x86_64-linux-ubuntu-20.04..."
	textInput.Focus()
	textInput.CharLimit = 100
	textInput.Width = 100

	textInput.TextStyle = focusedStyle
	textInput.PromptStyle = focusedStyle
	textInput.Cursor.Style = focusedStyle

	return &platformCreateModel{
		textInput: textInput,
		finished:  false,
		err:       nil,
		callbacks: callbacks,
	}
}

type platformCreateModel struct {
	textInput textinput.Model
	finished  bool
	err       error

	callbacks config.BuildEnvCallbacks
}

func (p platformCreateModel) Init() tea.Cmd {
	return textinput.Blink
}

func (p platformCreateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return p, tea.Quit

		case "esc":
			return MenuModel, nil

		case "enter":
			// Clean platform name.
			platformName := strings.TrimSpace(p.textInput.Value())
			platformName = strings.TrimSuffix(platformName, ".json")
			p.textInput.SetValue(platformName)

			if err := p.callbacks.OnCreatePlatform(p.textInput.Value()); err != nil {
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

func (p platformCreateModel) View() string {
	if p.finished {
		return config.PlatformCreated(p.textInput.Value())
	}

	if p.err != nil {
		return config.PlatformCreateFailed(p.textInput.Value(), p.err)
	}

	return fmt.Sprintf("\n%s\n\n%s\n\n%s\n",
		titleStyle.Width(50).Render("Please enter your platform's name:"),
		p.textInput.View(),
		actionBarStyle.Render("[enter -> execute | esc -> back | ctrl+c/q -> quit]"),
	)
}

func (p *platformCreateModel) Reset() {
	p.textInput.Reset()
	p.finished = false
	p.err = nil
}
