package ui

import (
	"fmt"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

func createUsageModel(goback func()) usageModel {
	toolchainPath, _ := filepath.Abs("cmake/buildenv.cmake")

	content := fmt.Sprintf("\nWelcome to buildenv.\n"+
		"-----------------------------------\n"+
		"This is a simple tool to manage your cross build environment.\n\n"+
		"How to use in cmake project: \n"+
		"\033[34moption1: set(CMAKE_TOOLCHAIN_FILE \"%s\")\033[0m\n"+
		"\033[34moption2: cmake .. -DCMAKE_TOOLCHAIN_FILE %s\033[0m\n\n"+
		"[press ctrl+c or q to quit]", toolchainPath, toolchainPath)
	return usageModel{content: content, goback: goback}
}

type usageModel struct {
	content string
	goback  func()
}

func (usageModel) Init() tea.Cmd {
	return nil
}

func (a usageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit

		case "esc":
			a.goback()
			return a, nil
		}
	}
	return a, nil
}

func (a usageModel) View() string {
	return a.content
}
