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
	flag.BoolVar(&i.init, "init", false, "init buildenv with config repo.")
	flag.StringVar(&i.confRepoUrl, "conf_repo_url", "", "set conf repo's url, it's used with -init.")
	flag.StringVar(&i.confRepoRef, "conf_repo_ref", "", "set conf repo's ref, it's used with -init.")
}

func (i *initConfCmd) listen() (handled bool) {
	if !i.init {
		return false
	}

	output, err := i.callbacks.OnInitBuildEnv(i.confRepoUrl, i.confRepoRef)
	if err != nil {
		fmt.Println(config.ConfigInitFailed(i.confRepoUrl, err))
		return
	}

	fmt.Println(output)
	fmt.Print(config.ConfigInitialized())

	return true
}
