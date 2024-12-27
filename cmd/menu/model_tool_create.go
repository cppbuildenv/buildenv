package menu

import (
	"buildenv/config"
	"buildenv/pkg/color"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func newToolCreateModel(callbacks config.BuildEnvCallbacks) *toolCreateModel {
	ti := textinput.New()
	ti.Placeholder = "your tool's name..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 100
	ti.TextStyle = styleImpl.focusedStyle
	ti.PromptStyle = styleImpl.focusedStyle
	ti.Cursor.Style = styleImpl.focusedStyle

	return &toolCreateModel{
		textInput: ti,
		callbacks: callbacks,
	}
}

type toolCreateModel struct {
	textInput textinput.Model
	created   bool
	err       error

	callbacks config.BuildEnvCallbacks
}

func (t toolCreateModel) Init() tea.Cmd {
	return textinput.Blink
}

func (t toolCreateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return t, tea.Quit

		case "esc":
			return MenuModel, nil

		case "enter":
			// Clean tool name.
			toolName := strings.TrimSpace(t.textInput.Value())
			toolName = strings.TrimSuffix(toolName, ".json")
			t.textInput.SetValue(toolName)

			if err := t.callbacks.OnCreateTool(t.textInput.Value()); err != nil {
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

func (t toolCreateModel) View() string {
	if t.created {
		return config.ToolCreated(t.textInput.Value())
	}

	if t.err != nil {
		return config.ToolCreateFailed(t.textInput.Value(), t.err)
	}

	return fmt.Sprintf("\n%s\n\n%s\n\n%s\n",
		color.Sprintf(color.Blue, "Please enter your tool's name: "),
		t.textInput.View(),
		color.Sprintf(color.Gray, "[esc -> back | ctrl+c/q -> quit]"),
	)
}

func (t *toolCreateModel) Reset() {
	t.textInput.Reset()
	t.created = false
	t.err = nil
}
