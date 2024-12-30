package menu

import (
	"buildenv/config"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func newToolCreateModel(callbacks config.BuildEnvCallbacks) *toolCreateModel {
	textInput := textinput.New()
	textInput.Placeholder = "your tool's name..."
	textInput.Focus()
	textInput.CharLimit = 100
	textInput.Width = 100
	textInput.TextStyle = focusedStyle
	textInput.PromptStyle = focusedStyle
	textInput.Cursor.Style = focusedStyle

	return &toolCreateModel{
		textInput: textInput,
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
			t.reset()
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
		titleStyle.Width(50).Render("Please enter your tool's name:"),
		t.textInput.View(),
		actionBarStyle.Render("[enter -> execute | esc -> back | ctrl+c/q -> quit]"),
	)
}

func (t *toolCreateModel) reset() {
	t.textInput.Reset()
	t.created = false
	t.err = nil
}
