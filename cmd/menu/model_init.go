package menu

import (
	"buildenv/config"
	"buildenv/pkg/color"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func newInitModel(callbacks config.BuildEnvCallbacks) *initModel {
	ti := textinput.New()
	ti.Placeholder = "buildenv's config repo url..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 100
	ti.TextStyle = styleImpl.focusedStyle
	ti.PromptStyle = styleImpl.focusedStyle
	ti.Cursor.Style = styleImpl.focusedStyle

	return &initModel{
		textInput: ti,
		callbacks: callbacks,
	}
}

type initModel struct {
	textInput   textinput.Model
	initialized bool
	err         error

	callbacks config.BuildEnvCallbacks
}

func (i initModel) Init() tea.Cmd {
	return textinput.Blink
}

func (i initModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return i, tea.Quit

		case "esc":
			return MenuModel, nil

		case "enter":
			// Clean config repo url.
			configUrl := strings.TrimSpace(i.textInput.Value())
			configUrl = strings.TrimSuffix(configUrl, ".json")
			i.textInput.SetValue(configUrl)

			if err := i.callbacks.OnInitBuildEnv(i.textInput.Value()); err != nil {
				i.err = err
				i.initialized = false
			} else {
				i.initialized = true
			}

			return i, tea.Quit
		}
	}

	var cmd tea.Cmd
	i.textInput, cmd = i.textInput.Update(msg)
	return i, cmd
}

func (i initModel) View() string {
	if i.initialized {
		return config.ConfigInitialized(i.textInput.Value())
	}

	if i.err != nil {
		return config.ConfigInitFailed(i.textInput.Value(), i.err)
	}

	return fmt.Sprintf("\n%s\n\n%s\n\n%s\n",
		color.Sprintf(color.Blue, "Please enter buildenv's config url: "),
		i.textInput.View(),
		color.Sprintf(color.Gray, "[esc -> back | ctrl+c/q -> quit]"),
	)
}

func (i *initModel) Reset() {
	i.textInput.Reset()
	i.initialized = false
	i.err = nil
}
