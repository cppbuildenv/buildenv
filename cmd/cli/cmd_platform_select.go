package cli

import (
	"buildenv/config"
	"flag"
	"fmt"
	"strings"
)

func newPlatformSelectCmd(callbacks config.BuildEnvCallbacks) *platformSelectCmd {
	return &platformSelectCmd{
		callbacks: callbacks,
	}
}

type platformSelectCmd struct {
	platformName string
	callbacks    config.BuildEnvCallbacks
}

func (p *platformSelectCmd) register() {
	flag.StringVar(&p.platformName, "select_platform", "", "select a platform as cross build environment.")
}

func (p *platformSelectCmd) listen() (handled bool) {
	if p.platformName == "" {
		return false
	}

	// Clean platform name.
	p.platformName = strings.TrimSpace(p.platformName)
	p.platformName = strings.TrimSuffix(p.platformName, ".json")

	if err := p.callbacks.OnSelectPlatform(p.platformName); err != nil {
		fmt.Print(config.PlatformSelectedFailed(p.platformName, err))
		return true
	}

	fmt.Print(config.PlatformSelected(p.platformName))
	return true
}
