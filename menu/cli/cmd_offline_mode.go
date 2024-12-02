package cli

import (
	"buildenv/config"
	"flag"
	"fmt"
)

func newOfflineCmd(callbacks config.PlatformCallbacks) *offlineCmd {
	return &offlineCmd{
		callbacks: callbacks,
	}
}

type offlineCmd struct {
	callbacks config.PlatformCallbacks
	offline   bool
}

func (o *offlineCmd) register() {
	flag.BoolVar(&o.offline, "offline", false, "offline mode")
}

func (o *offlineCmd) listen() (handled bool) {
	o.callbacks.SetOffline(o.offline)
	fmt.Print(config.SetOffline(o.offline))
	return true
}
