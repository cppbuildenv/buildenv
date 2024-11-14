package ui

import (
	"buildenv/config"
	"buildenv/console"
	"buildenv/pkg/color"
	"buildenv/pkg/io"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func newSyncConfigModel(goback func()) *syncConfigModel {
	content := fmt.Sprintf("\nClone or synch repo of conf.\n"+
		"-----------------------------------\n"+
		"%s.\n\n"+
		"%s",
		color.Sprintf(color.Blue, "This will create a buildenv.json if not exist, otherwise it'll checkout the latest conf repo with specified repo REF"),
		color.Sprintf(color.Gray, "[↵ -> execute | ctrl+c/q -> quit]"))

	return &syncConfigModel{
		content: content,
		goback:  goback,
	}
}

type syncConfigModel struct {
	content string
	output  string
	goback  func()
}

func (s syncConfigModel) Init() tea.Cmd {
	return nil
}

func (s syncConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return s, tea.Quit

		case "enter":
			if output, err := s.syncRepo(); err != nil {
				s.output = color.Sprintf(color.Red, "Error: %s", err.Error())
			} else {
				s.output = color.Sprintf(color.Blue, output) + "\n" + console.SyncSuccess(true)
			}
			return s, tea.Quit

		case "esc":
			s.goback()
			return s, nil
		}
	}
	return s, nil
}

func (s syncConfigModel) View() string {
	if s.output != "" {
		return s.output
	}

	return s.content
}

func (s syncConfigModel) syncRepo() (string, error) {
	// Create buildenv.json if not exist.
	confPath := filepath.Join(config.Dirs.WorkspaceDir, "buildenv.json")
	if !io.PathExists(confPath) {
		if err := os.MkdirAll(filepath.Dir(confPath), os.ModePerm); err != nil {
			return "", err
		}

		var buildenv config.BuildEnv
		buildenv.JobNum = runtime.NumCPU()

		bytes, err := json.MarshalIndent(buildenv, "", "    ")
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(confPath, []byte(bytes), os.ModePerm); err != nil {
			return "", err
		}

		return console.SyncSuccess(false), nil
	}

	// Sync conf repo with repo url.
	bytes, err := os.ReadFile(confPath)
	if err != nil {
		return "", err
	}

	// Unmarshall with buildenv.json.
	var buildenv config.BuildEnv
	if err := json.Unmarshal(bytes, &buildenv); err != nil {
		return "", err
	}

	// Sync repo.
	outputs, err := buildenv.SyncRepo(buildenv.ConfRepo, buildenv.ConfRepoRef)
	if err != nil {
		return "", err
	}

	return strings.Join(outputs, "\n"), nil
}
