package cli

import "flag"

func newBuildTypeCmd() *buildTypeCmd {
	return &buildTypeCmd{}
}

type buildTypeCmd struct {
	buildType string
}

func (b *buildTypeCmd) register() {
	flag.StringVar(&b.buildType, "build_type", "Release", "set build type and works with '--install' and '--setup', for example:./buildenv --build_type=Debug")
}
