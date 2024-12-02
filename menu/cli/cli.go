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
	silent         = newSilentModeCmd()
	buildType      = newBuildTypeCmd()
	ui             = newUICmd(config.PlatformCallbacksImpl)
	version        = newVersionCmd()
	sync           = newSyncConfigCmd()
	createPlatform = newCreatePlatformCmd()
	selectPlatform = newSelectPlatformCmd(config.PlatformCallbacksImpl)
	verify         = newVerifyCmd()
)
var commands = []reisterable{
	silent,
	buildType,
	ui,
	version,
	sync,
	createPlatform,
	selectPlatform,
	verify,
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
