package cli

import (
	"buildenv/config"
	"buildenv/console"
	"flag"
	"fmt"
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
		fmt.Printf(console.PlatformCreateFailed, c.platformName, err)
		return true
	}

	fmt.Printf(console.PlatformCreated, c.platformName)
	return true
}

func (c *createPlatformCmd) doCreate(platformName string) error {
	buildEnv := config.BuildEnv{}
	if err := buildEnv.Write(platformName); err != nil {
		return err
	}

	return nil
}
