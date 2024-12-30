package menu

import (
	"buildenv/config"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	repo_url = iota
	repo_ref
)

func newInitModel(callbacks config.BuildEnvCallbacks) *initModel {
	var inputs []textinput.Model = make([]textinput.Model, 2)

	inputs[repo_url] = textinput.New()
	inputs[repo_url].Placeholder = "git@192.169.x.x:buildenv/config/repo.git"
	inputs[repo_url].Focus()
	inputs[repo_url].CharLimit = 100
	inputs[repo_url].Width = 100
	inputs[repo_url].Prompt = ""
	// inputs[repor_url].Validate = ccnValidator

	inputs[repo_ref] = textinput.New()
	inputs[repo_ref].Placeholder = "master"
	inputs[repo_ref].CharLimit = 20
	inputs[repo_ref].Width = 20
	inputs[repo_ref].Prompt = ""
	// inputs[repo_ref].Validate = expValidator

	return &initModel{
		textInputs: inputs,
		focused:    0,
		finished:   false,
		output:     "",
		err:        nil,
		callbacks:  callbacks,
	}
}

type initModel struct {
	textInputs []textinput.Model
	focused    int
	finished   bool
	output     string
	err        error

	callbacks config.BuildEnvCallbacks
}

func (i initModel) Init() tea.Cmd {
	return textinput.Blink
}

func (i initModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd = make([]tea.Cmd, len(i.textInputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if i.focused == len(i.textInputs)-1 {
				repoUrl := i.textInputs[repo_url].Value()
				repoRef := i.textInputs[repo_ref].Value()
				output, err := i.callbacks.OnInitBuildEnv(repoUrl, repoRef)
				if err != nil {
					i.err = err
					i.output = output
					i.finished = false
				} else {
					i.finished = true
				}

				return i, tea.Quit
			}
			i.nextInput()

		case "shift+tab":
			i.prevInput()

		case "tab":
			i.nextInput()

		case "ctrl+c", "q":
			return i, tea.Quit

		case "esc":
			i.reset()
			return MenuModel, nil
		}

		for index := range i.textInputs {
			i.textInputs[index].Blur()
		}
		i.textInputs[i.focused].Focus()

	case errMsg:
		i.err = msg
		return i, nil
	}

	for index := range i.textInputs {
		i.textInputs[index], cmds[index] = i.textInputs[index].Update(msg)
	}
	return i, tea.Batch(cmds...)
}

func (i initModel) View() string {
	repoUrl := i.textInputs[repo_url].Value()

	if i.finished {
		return config.ConfigInitialized()
	}

	if i.err != nil {
		return config.ConfigInitFailed(repoUrl, i.err)
	}

	return fmt.Sprintf("\nInitializing buildenv:\n\n%s\n%s\n\n%s\n%s\n\n%s\n",
		inputStyle.Width(30).Render("Config repo url"),
		i.textInputs[repo_url].View(),
		inputStyle.Width(30).Render("Config repo ref"),
		i.textInputs[repo_ref].View(),
		actionBarStyle.Render("[enter -> execute | esc -> back | ctrl+c/q -> quit]"),
	) + "\n"
}

func (i *initModel) nextInput() {
	i.focused = (i.focused + 1) % len(i.textInputs)
}

func (i *initModel) prevInput() {
	i.focused--
	if i.focused < 0 { // Wrap around
		i.focused = len(i.textInputs) - 1
	}
}

func (i *initModel) reset() {
	i.err = nil
	i.focused = 0

	for index := range i.textInputs {
		i.textInputs[index].Blur()
		i.textInputs[index].Reset()
	}
}
