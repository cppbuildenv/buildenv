package interactive

import (
	"buildenv/config"
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func createPlatformCreateModel(callbacks config.PlatformCallbacks, goback func(this *platformCreateModel)) platformCreateModel {
	ti := textinput.New()
	ti.Placeholder = "for example: x86_64-linux-ubuntu-20.04..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 100
	ti.TextStyle = styleImpl.focusedStyle
	ti.PromptStyle = styleImpl.focusedStyle
	ti.Cursor.Style = styleImpl.focusedStyle

	return platformCreateModel{
		textInput: ti,
		callbacks: callbacks,
		goback:    goback,
		styles:    styleImpl,
	}
}

type platformCreateModel struct {
	textInput    textinput.Model
	platformName string
	styles       styles
	created      bool
	err          error

	callbacks config.PlatformCallbacks
	goback    func(this *platformCreateModel)
}

func (p platformCreateModel) Init() tea.Cmd {
	return textinput.Blink
}

func (p platformCreateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			p.goback(&p)
			return p, nil

		case "enter":
			filePath := filepath.Join(config.PlatformDir, p.textInput.Value()+".json")
			if err := p.callbacks.OnCreatePlatform(filePath); err != nil {
				p.err = err
				p.created = false
			} else {
				p.platformName = p.textInput.Value()
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
		return p.styles.resultTextStyle.Render("[✔] ---- platform create success:", p.platformName)
	}

	if p.err != nil {
		return p.styles.resultTextStyle.Render("[✘] ---- platform create failed:", p.err.Error())
	}

	return fmt.Sprintf("\n%s\n\n%s\n\n%s\n",
		p.styles.titleStyle.Render("Please input your platform name: "),
		p.textInput.View(),
		p.styles.helpStyle.Render("[esc/q back · ctrl+c quit]"),
	)
}

func (p *platformCreateModel) Reset() {
	p.textInput.Reset()
	p.platformName = ""
	p.created = false
	p.err = nil
}