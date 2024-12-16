package cli

import (
	"buildenv/config"
	"flag"
	"fmt"
	"path/filepath"
	"strings"
)

func newPlatformCreateCmd() *platformCreateCmd {
	return &platformCreateCmd{}
}

type platformCreateCmd struct {
	platformName string
}

func (p *platformCreateCmd) register() {
	flag.StringVar(&p.platformName, "create_platform", "", "create a new platform with template.")
}

func (p *platformCreateCmd) listen() (handled bool) {
	if p.platformName == "" {
		return false
	}

	// Clean platform name.
	p.platformName = strings.TrimSpace(p.platformName)
	p.platformName = strings.TrimSuffix(p.platformName, ".json")

	if err := p.doCreate(p.platformName); err != nil {
		fmt.Print(config.PlatformCreateFailed(p.platformName, err))
		return true
	}

	fmt.Print(config.PlatformCreated(p.platformName))
	return true
}

func (p *platformCreateCmd) doCreate(platformName string) error {
	platformPath := filepath.Join(config.Dirs.PlatformsDir, platformName+".json")

	var platform config.Platform
	if err := platform.Write(platformPath); err != nil {
		return err
	}

	return nil
}
