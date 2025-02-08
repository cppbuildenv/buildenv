package cli

import (
	"buildenv/config"
	"buildenv/pkg/fileio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func newPurgeCmd() *purgeCmd {
	return &purgeCmd{}
}

type purgeCmd struct {
	removeCmd
}

func (p *purgeCmd) register() {
	flag.StringVar(&p.purge, "purge", "", "remove a third-party from installed dir and package dir, "+
		" for example: ./buildenv --purge glog@v0.6.0, you can also call with '--dev' to remove a dev third-party.")

	// Remove
	p.removeCmd.portRemoved = func(ctx config.Context, port config.Port) error {
		var folderName string
		if port.AsDev {
			folderName = port.NameVersion()
		} else {
			folderName = fmt.Sprintf("%s^%s^%s^%s",
				port.NameVersion(),
				ctx.Platform().Name,
				ctx.Project().Name,
				ctx.BuildType(),
			)
		}

		// Remove port's package files.
		packageDir := filepath.Join(config.Dirs.WorkspaceDir, "packages", folderName)
		if err := os.RemoveAll(packageDir); err != nil {
			return fmt.Errorf("cannot remove package files: %s", err)
		}

		// Try remove parent folder if it's empty.
		if err := fileio.RemoveFolderRecursively(filepath.Dir(packageDir)); err != nil {
			return fmt.Errorf("cannot remove parent folder: %s", err)
		}

		return nil
	}
}

func (p *purgeCmd) listen() (handled bool) {
	return p.removeCmd.listen()
}
