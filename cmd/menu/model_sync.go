package menu

import (
	"buildenv/config"
	"buildenv/pkg/color"
	"buildenv/pkg/io"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

func newSyncModel() *initModel {
	content := fmt.Sprintf("\nClone or synch repo of conf.\n"+
		"-----------------------------------\n"+
		"%s.\n\n"+
		"%s",
		color.Sprintf(color.Blue, "This will create a buildenv.json if not exist, otherwise it'll checkout the latest commit."),
		color.Sprintf(color.Gray, "[↵ -> execute | ctrl+c/q -> quit]"))

	return &initModel{
		content: content,
	}
}

type initModelsyncModel struct {
	content string
}

func (s initModel) Init() tea.Cmd {
	return nil
}

func (s initModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return s, tea.Quit

		case "enter":
			if output, err := s.initModel(); err != nil {
				s.content += "\r" + color.Sprintf(color.Red, err.Error())
			} else {
				s.content += "\r" + output + "\n" + config.SyncSuccess(true)
			}
			return s, tea.Quit

		case "esc":
			return MenuModel, nil
		}
	}
	return s, nil
}

func (s initModel) View() string {
	return s.content
}

func (s initModel) syncRepo() (string, error) {
	// In cli ui mode, buildType is always `Release`.
	buildenv := config.NewBuildEnv("Release")

	// Create buildenv.json if not exist.
	confPath := filepath.Join(config.Dirs.WorkspaceDir, "buildenv.json")
	if !io.PathExists(confPath) {
		if err := os.MkdirAll(filepath.Dir(confPath), os.ModePerm); err != nil {
			return "", err
		}

		bytes, err := json.MarshalIndent(buildenv, "", "    ")
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(confPath, []byte(bytes), os.ModePerm); err != nil {
			return "", err
		}

		return config.SyncSuccess(false), nil
	}

	// Sync conf repo with repo url.
	bytes, err := os.ReadFile(confPath)
	if err != nil {
		return "", err
	}

	// Unmarshall with buildenv.json.
	if err := json.Unmarshal(bytes, &buildenv); err != nil {
		return "", err
	}

	// Sync repo.
	return buildenv.Synchronize(buildenv.ConfRepoUrl, buildenv.ConfRepoRef)
}
