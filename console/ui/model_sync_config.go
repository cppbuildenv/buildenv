package ui

import (
	"buildenv/config"
	"buildenv/console"
	"buildenv/pkg/color"
	"buildenv/pkg/io"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
)

func newSyncConfigModel(goback func()) *syncConfigModel {
	content := fmt.Sprintf("\nClone or synch repo of conf.\n"+
		"-----------------------------------\n"+
		"%s.\n\n"+
		"%s",
		color.Sprintf(color.Blue, "This will create a buildenv.json if not exist, otherwise it'll checkout the latest conf repo with specified repo REF"),
		color.Sprintf(color.Gray, "[↵ Execute | ctrl+c/q Quit]"))

	return &syncConfigModel{
		content: content,
		goback:  goback,
	}
}

type syncConfigModel struct {
	content string
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
			s.syncRepo()
			return s, tea.Quit

		case "esc":
			s.goback()
			return s, nil
		}
	}
	return s, nil
}

func (s syncConfigModel) View() string {
	return s.content
}

func (s syncConfigModel) syncRepo() {
	// Create buildenv.json if not exist.
	confPath := filepath.Join(config.Dirs.WorkspaceDir, "buildenv.json")
	if !io.PathExists(confPath) {
		if err := os.MkdirAll(filepath.Dir(confPath), os.ModePerm); err != nil {
			log.Fatal(err)
		}

		var buildenv config.BuildEnv
		buildenv.JobNum = runtime.NumCPU()

		bytes, err := json.MarshalIndent(buildenv, "", "    ")
		if err != nil {
			fmt.Print(console.SyncFailed(err))
			os.Exit(1)
		}
		if err := os.WriteFile(confPath, []byte(bytes), os.ModePerm); err != nil {
			fmt.Print(console.SyncFailed(err))
			os.Exit(1)
		}

		fmt.Print(console.SyncSuccess(false))
		return
	}

	// Sync conf repo with repo url.
	bytes, err := os.ReadFile(confPath)
	if err != nil {
		fmt.Print(console.SyncFailed(err))
		os.Exit(1)
	}

	// Unmarshall with buildenv.json.
	var buildenv config.BuildEnv
	if err := json.Unmarshal(bytes, &buildenv); err != nil {
		fmt.Print(console.SyncFailed(err))
		os.Exit(1)
	}

	// Sync repo.
	if err := buildenv.SyncRepo(buildenv.ConfRepo, buildenv.ConfRepoRef); err != nil {
		fmt.Print(console.SyncFailed(err))
		os.Exit(1)
	}
}
