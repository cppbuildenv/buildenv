package cli

import (
	"buildenv/config"
)

type command struct {
	Name        string
	Description string
	Handler     func(callbacks config.BuildEnvCallbacks)
}

var Commands = []command{
	{
		Name:        "init",
		Description: "Initialize buildenv.",
		Handler:     handleInitialize,
	},
	{
		Name:        "setup",
		Description: "Setup buildenv for selected platform and project.",
		Handler:     handleSetup,
	},
	{
		Name:        "install",
		Description: "Install a third-party library.",
		Handler:     handleInstall,
	},
	{
		Name:        "remove",
		Description: "Remove an installed third-party library.",
		Handler:     handleRemove,
	},
	{
		Name:        "create",
		Description: "Create platform, project, tool or port.",
		Handler:     handleCreate,
	},
	{
		Name:        "select",
		Description: "Select platform or platform.",
		Handler:     handleSelect,
	},
	{
		Name:        "integrate",
		Description: "Integrate buildenv so can call it anywhere.",
		Handler:     handleIntegrate,
	},
	{
		Name:        "about",
		Description: "About buildenv and usage.",
		Handler:     handleAbout,
	},
}
