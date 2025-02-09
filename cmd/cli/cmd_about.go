package cli

import (
	"buildenv/config"
	"flag"
	"fmt"
)

var Version string // for example: `1.0.0`

func newAboutCmd(callbacks config.BuildEnvCallbacks) *aboutCmd {
	return &aboutCmd{
		callbacks: callbacks,
	}
}

type aboutCmd struct {
	callbacks config.BuildEnvCallbacks
	about     bool
}

func (a *aboutCmd) register() {
	flag.BoolVar(&a.about, "about", false, "about buildenv and how to use it, for example: ./buildenv --about")
}

func (a *aboutCmd) listen() (handled bool) {
	if a.about {
		fmt.Print(a.callbacks.About(Version))
		return true
	}

	return false
}

func handleAbout(callbacks config.BuildEnvCallbacks) {
	cmd := flag.NewFlagSet("about", flag.ExitOnError)

	cmd.Usage = func() {
		fmt.Print("Usage: buildenv about\n\n")
		fmt.Println("options:")
		cmd.PrintDefaults()
	}

	fmt.Print(callbacks.About(Version))
}
