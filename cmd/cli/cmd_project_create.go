package cli

import (
	"buildenv/config"
	"flag"
	"fmt"
	"path/filepath"
	"strings"
)

func newProjectCreateCmd() *projectCreateCmd {
	return &projectCreateCmd{}
}

type projectCreateCmd struct {
	projectName string
}

func (p *projectCreateCmd) register() {
	flag.StringVar(&p.projectName, "create_project", "", "create a new project with template.")
}

func (p *projectCreateCmd) listen() (handled bool) {
	if p.projectName == "" {
		return false
	}

	// Clean project name.
	p.projectName = strings.TrimSpace(p.projectName)
	p.projectName = strings.TrimSuffix(p.projectName, ".json")

	if err := p.doCreate(p.projectName); err != nil {
		fmt.Print(config.ProjectCreateFailed(p.projectName, err))
		return true
	}

	fmt.Print(config.ProjectCreated(p.projectName))
	return true
}

func (p *projectCreateCmd) doCreate(projectName string) error {
	projectPath := filepath.Join(config.Dirs.ProjectsDir, projectName+".json")

	var project config.Project
	if err := project.Write(projectPath); err != nil {
		return err
	}

	return nil
}
