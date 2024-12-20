package cli

import (
	"buildenv/config"
	"buildenv/pkg/io"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func newSyncCmd() *syncCmd {
	return &syncCmd{}
}

type syncCmd struct {
	sync bool
}

func (s *syncCmd) register() {
	flag.BoolVar(&s.sync, "sync", false, "init or sync buildenv's config repo.")
}

func (s *syncCmd) listen() (handled bool) {
	if !s.sync {
		return false
	}

	buildenv := config.NewBuildEnv(buildType.buildType)

	// Create buildenv.json if not exist.
	handled = true
	confPath := filepath.Join(config.Dirs.WorkspaceDir, "buildenv.json")
	if !io.PathExists(confPath) {
		if err := os.MkdirAll(filepath.Dir(confPath), os.ModePerm); err != nil {
			fmt.Print(config.SyncFailed(err))
			return
		}

		bytes, err := json.MarshalIndent(buildenv, "", "    ")
		if err != nil {
			fmt.Print(config.SyncFailed(err))
			return
		}
		if err := os.WriteFile(confPath, []byte(bytes), os.ModePerm); err != nil {
			fmt.Print(config.SyncFailed(err))
			return
		}

		fmt.Print(config.SyncSuccess(false))
		return
	}

	// Sync conf repo with repo url.
	bytes, err := os.ReadFile(confPath)
	if err != nil {
		fmt.Print(config.SyncFailed(err))
		return
	}

	// Unmarshall with buildenv.json.
	if err := json.Unmarshal(bytes, &buildenv); err != nil {
		fmt.Print(config.SyncFailed(err))
		return
	}

	// Sync repo.
	output, err := buildenv.Synchronize(buildenv.ConfRepoUrl, buildenv.ConfRepoRef)
	if err != nil {
		fmt.Print(config.SyncFailed(err))
		return
	}

	fmt.Println(output)
	fmt.Print(config.SyncSuccess(true))

	return
}
