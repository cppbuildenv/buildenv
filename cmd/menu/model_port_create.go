package menu

import (
	"buildenv/config"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func newPortCreateModel(callbacks config.BuildEnvCallbacks) *portCreateModel {
	textInput := textinput.New()
	textInput.Placeholder = "your port's name..."
	textInput.Focus()
	textInput.CharLimit = 100
	textInput.Width = 100
	textInput.TextStyle = focusedStyle
	textInput.PromptStyle = focusedStyle
	textInput.Cursor.Style = focusedStyle

	return &portCreateModel{
		textInput: textInput,
		finished:  false,
		err:       nil,
		callbacks: callbacks,
	}
}

type portCreateModel struct {
	textInput textinput.Model
	finished  bool
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
			t.reset()
			return MenuModel, nil

		case "enter":
			// Clean port name.
			toolName := strings.TrimSpace(t.textInput.Value())
			toolName = strings.TrimSuffix(toolName, ".json")
			t.textInput.SetValue(toolName)

			if err := t.callbacks.OnCreatePort(t.textInput.Value()); err != nil {
				t.err = err
				t.finished = false
			} else {
				t.finished = true
			}

			return t, tea.Quit
		}
	}

	var cmd tea.Cmd
	t.textInput, cmd = t.textInput.Update(msg)
	return t, cmd
}

func (t portCreateModel) View() string {
	if t.finished {
		return config.PortCreated(t.textInput.Value())
	}

	if t.err != nil {
		return config.PortCreateFailed(t.textInput.Value(), t.err)
	}

	return fmt.Sprintf("\n%s\n\n%s\n\n%s\n",
		titleStyle.Width(50).Render("Please enter your port's name:"),
		t.textInput.View(),
		actionBarStyle.Render("[enter -> execute | esc -> back | ctrl+c/q -> quit]"),
	)
}

func (t *portCreateModel) reset() {
	t.textInput.Reset()
	t.finished = false
	t.err = nil
}
