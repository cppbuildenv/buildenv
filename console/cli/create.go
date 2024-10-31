package cli

import (
	"buildenv/config"
	"flag"
	"fmt"
)

type createCmd struct {
	platformName string
}

func (cmd *createCmd) register() {
	flag.StringVar(&cmd.platformName, "c", "", "create a new platform")
	flag.StringVar(&cmd.platformName, "create", "", "create a new platform")
}

func (cmd *createCmd) listen() (handled bool) {
	if cmd.platformName == "" {
		return false
	}

	if err := cmd.doCreate(cmd.platformName); err != nil {
		fmt.Printf("[✘] ---- platform create failed: %s", err)
		return true
	}

	fmt.Printf("[✔] ---- platform create success: %s\n", cmd.platformName)
	return true
}

func (cmd *createCmd) doCreate(name string) error {
	buildEnv := config.BuildEnv{}
	if err := buildEnv.Write(name, force.force); err != nil {
		return err
	}

	return nil
}
