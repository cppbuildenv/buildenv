package cmd

import (
	"buildenv/config"
	"flag"
	"fmt"
	"log"
)

type createCmd struct {
	platformName string
}

func (cmd *createCmd) register() {
	flag.StringVar(&cmd.platformName, "c", "", "create buildenv")
	flag.StringVar(&cmd.platformName, "create", "", "create buildenv")
}

func (cmd *createCmd) listen() (quit bool) {
	if cmd.platformName == "" {
		return false
	}

	if err := cmd.doCreate(cmd.platformName); err != nil {
		log.Printf("[cfg/platform/%s]: %s", cmd.platformName, err)
		return true
	}

	log.Printf("[cfg/platform/%s]: created successfully...", cmd.platformName)
	return true
}

func (cmd *createCmd) doCreate(name string) error {
	buildEnv := config.BuildEnv{}
	if err := buildEnv.Write(name, force.force); err != nil {
		return fmt.Errorf("failed to write cfg/buildenv.json: %w", err)
	}

	return nil
}
