package ui

import (
	"fmt"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

func createUsageModel(goback func()) usageModel {
	toolchainPath, _ := filepath.Abs("script/buildenv.cmake")
	environmentPath, _ := filepath.Abs("script/buildenv.sh")

	content := fmt.Sprintf("\nWelcome to buildenv.\n"+
		"-----------------------------------\n"+
		"This is a simple tool to manage your cross build environment.\n\n"+
		"1. How to use in cmake project: \n"+
		"option1: \033[34mset(CMAKE_TOOLCHAIN_FILE \"%s\")\033[0m\n"+
		"option2: \033[34mcmake .. -DCMAKE_TOOLCHAIN_FILE=%s\033[0m\n\n"+
		"2. How to use in makefile project: \n"+
		"\033[34msource %s\033[0m\n\n"+
		"[press ctrl+c or q to quit]",
		toolchainPath,
		toolchainPath,
		environmentPath,
	)
	return usageModel{content: content, goback: goback}
}

type usageModel struct {
	content string
	goback  func()
}

func (usageModel) Init() tea.Cmd {
	return nil
}

func (u usageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return u, tea.Quit

		case "esc":
			u.goback()
			return u, nil
		}
	}
	return u, nil
}

func (u usageModel) View() string {
	return u.content
}
