package cli

import (
	"buildenv/config"
	"flag"
	"fmt"
)

func newSelectPlatformCmd(callbacks config.BuildEnvCallbacks) *selectPlatformCmd {
	return &selectPlatformCmd{
		callbacks: callbacks,
	}
}

type selectPlatformCmd struct {
	platformName string
	callbacks    config.BuildEnvCallbacks
}

func (s *selectPlatformCmd) register() {
	flag.StringVar(&s.platformName, "select_platform", "", "select a platform as cross build environment.")
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
