package cli

import (
	"buildenv/config"
	"flag"
	"fmt"
)

func newInitConfCmd(callbacks config.BuildEnvCallbacks) *initConfCmd {
	return &initConfCmd{
		callbacks: callbacks,
	}
}

type initConfCmd struct {
	init        bool
	confRepoUrl string
	confRepoRef string
	callbacks   config.BuildEnvCallbacks
}

func (i *initConfCmd) register() {
	flag.BoolVar(&i.init, "init", false, "init buildenv with config repo, works with '--conf_repo_url' and '--conf_repo_ref' to set repo url and ref.")
	flag.StringVar(&i.confRepoUrl, "conf_repo_url", "", "set conf repo's url and wotks with '--init'.")
	flag.StringVar(&i.confRepoRef, "conf_repo_ref", "master", "set conf repo's ref and works with '--init'.")
}

func (i *initConfCmd) listen() (handled bool) {
	if !i.init {
		return false
	}

	output, err := i.callbacks.OnInitBuildEnv(i.confRepoUrl, i.confRepoRef)
	if err != nil {
		config.PrintError(err, "failed to init buildenv with %s.", i.confRepoUrl)
		return
	}

	fmt.Println(output)
	config.PrintSuccess("init buildenv successfully.")

	return true
}
