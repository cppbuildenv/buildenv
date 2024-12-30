package cli

import (
	"buildenv/config"
	"flag"
	"strings"
)

func newPortCreateCmd(callbacks config.BuildEnvCallbacks) *portCreateCmd {
	return &portCreateCmd{
		callbacks: callbacks,
	}
}

type portCreateCmd struct {
	portNameVersion string
	callbacks       config.BuildEnvCallbacks
}

func (p *portCreateCmd) register() {
	flag.StringVar(&p.portNameVersion, "create_port", "", "create a new port with template.")
}

func (p *portCreateCmd) listen() (handled bool) {
	if p.portNameVersion == "" {
		return false
	}

	// Clean port name.
	p.portNameVersion = strings.TrimSpace(p.portNameVersion)
	p.portNameVersion = strings.TrimSuffix(p.portNameVersion, ".json")

	if err := p.callbacks.OnCreatePort(p.portNameVersion); err != nil {
		config.PrintError(err, "%s could not be created.", p.portNameVersion)
		return true
	}

	config.PrintSuccess("%s is created but need to config it later.", p.portNameVersion)
	return true
}
