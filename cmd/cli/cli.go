package cli

import (
	"buildenv/config"
	"flag"
	"runtime"
)

type reisterable interface {
	register()
}

type responsible interface {
	reisterable
	listen() (handled bool)
}

var (
	buildType      = newBuildTypeCmd()
	initConfig     = newInitConfCmd(config.Callbacks)
	platformCreate = newPlatformCreateCmd()
	platformSelect = newPlatformSelectCmd(config.Callbacks)
	projectCreate  = newProjectCreateCmd()
	projectSelect  = newProjectSelectCmd(config.Callbacks)
	toolCreate     = newToolCreateCmd(config.Callbacks)
	portCreate     = newPortCreateCmd(config.Callbacks)
	verify         = newVerifyCmd()
	install        = newInstallCmd()
	uninstall      = newUninstallCmd()
	recursive      = newRecursiveCmd()
	about          = newAboutCmd(config.Callbacks)
)
var commands = []reisterable{
	buildType,
	initConfig,
	platformCreate,
	platformSelect,
	projectCreate,
	projectSelect,
	toolCreate,
	portCreate,
	verify,
	install,
	uninstall,
	recursive,
	about,
}

func BuildType() string {
	return buildType.buildType
}

// Listen listen commands input
func Listen() bool {
	// `integrate` is supported in unix like system only.
	if runtime.GOOS == "linux" {
		integrate := newIntegrateCmd()
		commands = append(commands, integrate)
	}

	// Read command with flag
	for i := 0; i < len(commands); i++ {
		commands[i].register()
	}
	flag.Parse()

	// Handle commands
	for i := 0; i < len(commands); i++ {
		if cmd, ok := commands[i].(responsible); ok {
			if cmd.listen() {
				return true
			}
		}
	}

	return false
}
