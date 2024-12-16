package menu

import (
	"buildenv/config"
	"buildenv/pkg/color"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func newPlatformCreateModel(callbacks config.BuildEnvCallbacks) *platformCreateModel {
	ti := textinput.New()
	ti.Placeholder = "for example: x86_64-linux-ubuntu-20.04..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 100
	ti.TextStyle = styleImpl.focusedStyle
	ti.PromptStyle = styleImpl.focusedStyle
	ti.Cursor.Style = styleImpl.focusedStyle

	return &platformCreateModel{
		textInput: ti,
		callbacks: callbacks,
	}
}

type platformCreateModel struct {
	textInput textinput.Model
	created   bool
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

func (p platformCreateModel) View() string {
	if p.created {
		return config.PlatformCreated(p.textInput.Value())
	}

	if p.err != nil {
		return config.PlatformCreateFailed(p.textInput.Value(), p.err)
	}

	return fmt.Sprintf("\n%s\n\n%s\n\n%s\n",
		color.Sprintf(color.Blue, "Please enter your platform name: "),
		p.textInput.View(),
		color.Sprintf(color.Gray, "[esc -> back | ctrl+c/q -> quit]"),
	)
}

func (p *platformCreateModel) Reset() {
	p.textInput.Reset()
	p.created = false
	p.err = nil
}
