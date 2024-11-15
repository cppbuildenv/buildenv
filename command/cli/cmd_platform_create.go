package cli

import (
	"buildenv/command"
	"buildenv/config"
	"flag"
	"fmt"
	"path/filepath"
)

func newCreatePlatformCmd() *createPlatformCmd {
	return &createPlatformCmd{}
}

type createPlatformCmd struct {
	platformName string
}

func (c *createPlatformCmd) register() {
	flag.StringVar(&c.platformName, "create_platform", "", "create a new platform")
}

func (c *createPlatformCmd) listen() (handled bool) {
	if c.platformName == "" {
		return false
	}

	if err := c.doCreate(c.platformName); err != nil {
		fmt.Print(command.PlatformCreateFailed(c.platformName, err))
		return true
	}

	fmt.Print(command.PlatformCreated(c.platformName))
	return true
}

func (c *createPlatformCmd) doCreate(platformName string) error {
	platformPath := filepath.Join(config.Dirs.PlatformDir, platformName+".json")

	var platform config.Platform
	if err := platform.Write(platformPath); err != nil {
		return err
	}

	return nil
}
