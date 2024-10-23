package cmd

import (
	"buildenv/pkg"
	"flag"
	"fmt"
	"log"
)

type createCmd struct {
	filename string
}

func (cmd *createCmd) register() {
	flag.StringVar(&cmd.filename, "create", "", "create buildenv")
}

func (cmd *createCmd) listen() (exit bool) {
	if cmd.filename == "" {
		return false
	}

	if err := cmd.create(cmd.filename); err != nil {
		log.Println(err)
	} else {
		log.Printf("[buildenv/%v.json] was successfully created with default template, please fill it...", cmd.filename)
	}
	return true
}

func (cmd *createCmd) create(filename string) error {
	buildEnv := pkg.BuildEnv{}
	if err := buildEnv.Write(filename); err != nil {
		return fmt.Errorf("failed to write buildenv %v", filename)
	}

	return nil
}
