package menu

import (
	"buildenv/config"
	"buildenv/pkg/color"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func newPortCreateModel(callbacks config.BuildEnvCallbacks) *portCreateModel {
	ti := textinput.New()
	ti.Placeholder = "your port's name..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 100
	ti.TextStyle = styleImpl.focusedStyle
	ti.PromptStyle = styleImpl.focusedStyle
	ti.Cursor.Style = styleImpl.focusedStyle

	return &portCreateModel{
		textInput: ti,
		callbacks: callbacks,
	}
}

type portCreateModel struct {
	textInput textinput.Model
	created   bool
	err       error

	callbacks config.BuildEnvCallbacks
}

func (t portCreateModel) Init() tea.Cmd {
	return textinput.Blink
}

func (t portCreateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return t, tea.Quit

		case "esc":
			return MenuModel, nil

		case "enter":
			// Clean port name.
			toolName := strings.TrimSpace(t.textInput.Value())
			toolName = strings.TrimSuffix(toolName, ".json")
			t.textInput.SetValue(toolName)

			if err := t.callbacks.OnCreatePort(t.textInput.Value()); err != nil {
				t.err = err
				t.created = false
			} else {
				t.created = true
			}

			return t, tea.Quit
		}
	}

	var cmd tea.Cmd
	t.textInput, cmd = t.textInput.Update(msg)
	return t, cmd
}

func (t portCreateModel) View() string {
	if t.created {
		return config.PortCreated(t.textInput.Value())
	}

	if t.err != nil {
		return config.PortCreateFailed(t.textInput.Value(), t.err)
	}

	return fmt.Sprintf("\n%s\n\n%s\n\n%s\n",
		color.Sprintf(color.Blue, "Please enter your port's name: "),
		t.textInput.View(),
		color.Sprintf(color.Gray, "[esc -> back | ctrl+c/q -> quit]"),
	)
}

func (t *portCreateModel) Reset() {
	t.textInput.Reset()
	t.created = false
	t.err = nil
}
