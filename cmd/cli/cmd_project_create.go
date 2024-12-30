package cli

import (
	"buildenv/config"
	"flag"
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
		config.PrintSuccess("%s could not be created.", p.projectName)
		return true
	}

	config.PrintSuccess("%s is created but need to config it later.", p.projectName)
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
