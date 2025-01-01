package cli

import (
	"buildenv/config"
	"flag"
	"strings"
)

func newProjectSelectCmd(callbacks config.BuildEnvCallbacks) *projectSelectCmd {
	return &projectSelectCmd{
		callbacks: callbacks,
	}
}

type projectSelectCmd struct {
	projectName string
	callbacks   config.BuildEnvCallbacks
}

func (p *projectSelectCmd) register() {
	flag.StringVar(&p.projectName, "select_project", "", "select a project as current project.")
}

func (p *projectSelectCmd) listen() (handled bool) {
	if p.projectName == "" {
		return false
	}

	// Clean project name.
	p.projectName = strings.TrimSpace(p.projectName)
	p.projectName = strings.TrimSuffix(p.projectName, ".json")

	if err := p.callbacks.OnSelectProject(p.projectName); err != nil {
		config.PrintError(err, "failed to select project: %s.", p.projectName)
		return true
	}

	config.PrintSuccess("buildenv is ready for project: %s.", p.projectName)
	return true
}
