package cli

import (
	"buildenv/config"
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

func (s *selectPlatformCmd) register() {
	flag.StringVar(&s.platformName, "select_platform", "", "select a platform as build environment")
}

func (s *selectPlatformCmd) listen() (handled bool) {
	if s.platformName == "" {
		return false
	}

	if err := s.callbacks.OnSelectPlatform(s.platformName); err != nil {
		fmt.Print(config.PlatformSelectedFailed(s.platformName, err))
		return true
	}

	fmt.Print(config.PlatformSelected(s.platformName))
	return true
}
