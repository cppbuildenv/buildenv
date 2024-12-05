package cli

import (
	"buildenv/config"
	"flag"
	"fmt"
)

func newAboutCmd(callbacks config.PlatformCallbacks) *aboutCmd {
	return &aboutCmd{
		callbacks: callbacks,
	}
}

type aboutCmd struct {
	callbacks config.PlatformCallbacks
	about     bool
}

func (a *aboutCmd) register() {
	flag.BoolVar(&a.about, "about", false, "about and usage")
}

func (a *aboutCmd) listen() (handled bool) {
	if a.about {
		fmt.Print(a.callbacks.About())
		return true
	}

	return false
}
