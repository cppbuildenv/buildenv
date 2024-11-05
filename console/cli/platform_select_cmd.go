package cli

import (
	"buildenv/config"
	"buildenv/console"
	"flag"
	"fmt"
	"path/filepath"
)

func newSelectPlatformCmd(platformDir string, callbacks config.PlatformCallbacks) *selectPlatformCmd {
	return &selectPlatformCmd{
		platformDir: platformDir,
		callbacks:   callbacks,
	}
}

type selectPlatformCmd struct {
	platformName string
	platformDir  string
	callbacks    config.PlatformCallbacks
}

func (p *selectPlatformCmd) register() {
	flag.StringVar(&p.platformName, "select_platform", "", "select a platform as build target platform")
}

func (s *selectPlatformCmd) listen() (handled bool) {
	if s.platformName == "" {
		return false
	}

	filePath := filepath.Join(config.PlatformsDir, s.platformName+".json")
	if err := s.callbacks.OnSelectPlatform(filePath); err != nil {
		fmt.Printf(console.PlatformSelectedFailed, s.platformName, err)
		return true
	}

	fmt.Printf(console.PlatformSelected, s.platformName)
	return true
}
