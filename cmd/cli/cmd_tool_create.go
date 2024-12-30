package cli

import (
	"buildenv/config"
	"flag"
	"strings"
)

func newToolCreateCmd(callbacks config.BuildEnvCallbacks) *toolCreateCmd {
	return &toolCreateCmd{
		callbacks: callbacks,
	}
}

type toolCreateCmd struct {
	toolName  string
	callbacks config.BuildEnvCallbacks
}

func (t *toolCreateCmd) register() {
	flag.StringVar(&t.toolName, "create_tool", "", "create a new tool with template.")
}

func (t *toolCreateCmd) listen() (handled bool) {
	if t.toolName == "" {
		return false
	}

	// Clean tool name.
	t.toolName = strings.TrimSpace(t.toolName)
	t.toolName = strings.TrimSuffix(t.toolName, ".json")

	if err := t.callbacks.OnCreateTool(t.toolName); err != nil {
		config.PrintError(err, "%s could not be created.", t.toolName)
		return true
	}

	config.PrintSuccess(" %s is created but need to config it later.", t.toolName)
	return true
}
