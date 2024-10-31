package cli

import (
	"buildenv/config"
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
	flag.StringVar(&c.platformName, "cp", "", "create a new platform")
	flag.StringVar(&c.platformName, "create_platform", "", "create a new platform")
}

func (c *createPlatformCmd) listen() (handled bool) {
	if c.platformName == "" {
		return false
	}

	if err := c.doCreate(c.platformName); err != nil {
		fmt.Printf("[✘] ---- platform create failed: %s", err)
		return true
	}

	fmt.Printf("[✔] ---- platform create success: %s\n", c.platformName)
	return true
}

func (c *createPlatformCmd) doCreate(name string) error {
	buildEnv := config.BuildEnv{}
	if err := buildEnv.Write(name, force.force); err != nil {
		return err
	}

	return nil
}
