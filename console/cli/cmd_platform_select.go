package cli

import (
	"buildenv/config"
	"buildenv/console"
	"flag"
	"fmt"
)

func newSelectPlatformCmd(callbacks config.PlatformCallbacks) *selectPlatformCmd {
	return &selectPlatformCmd{
		callbacks: callbacks,
	}
}

type selectPlatformCmd struct {
	platformName string
	callbacks    config.PlatformCallbacks
}

func (p *selectPlatformCmd) register() {
	flag.StringVar(&p.platformName, "select_platform", "", "select a platform as build target platform")
}

func (s *selectPlatformCmd) listen() (handled bool) {
	if s.platformName == "" {
		return false
	}

	if err := s.callbacks.OnSelectPlatform(s.platformName); err != nil {
		fmt.Print(console.PlatformSelectedFailed(s.platformName, err))
		return true
	}

	fmt.Print(console.PlatformSelected(s.platformName))
	return true
}