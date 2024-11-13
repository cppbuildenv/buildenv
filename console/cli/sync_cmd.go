package cli

import (
	"buildenv/config"
	"buildenv/console"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

func newSyncCmd() *syncCmd {
	return &syncCmd{}
}

type syncCmd struct {
	sync bool
}

func (s *syncCmd) register() {
	flag.BoolVar(&s.sync, "sync", false, "create buildenv.json or sync conf repo")
}

func (s *syncCmd) listen() (handled bool) {
	if !s.sync {
		return false
	}

	// Create buildenv.json if not exist.
	confPath := filepath.Join(config.Dirs.WorkspaceDir, "buildenv.json")
	if !pathExists(confPath) {
		if err := os.MkdirAll(filepath.Dir(confPath), os.ModePerm); err != nil {
			log.Fatal(err)
		}

		var buildenv config.BuildEnv
		buildenv.JobNum = runtime.NumCPU()

		bytes, err := json.MarshalIndent(buildenv, "", "    ")
		if err != nil {
			fmt.Print(console.SyncFailed(err))
			os.Exit(1)
		}
		if err := os.WriteFile(confPath, []byte(bytes), os.ModePerm); err != nil {
			fmt.Print(console.SyncFailed(err))
			os.Exit(1)
		}

		fmt.Print(console.SyncSuccess(false))
		return false
	}

	// Sync conf repo with repo url.
	bytes, err := os.ReadFile(confPath)
	if err != nil {
		fmt.Print(console.SyncFailed(err))
		os.Exit(1)
	}

	var buildenv config.BuildEnv
	if err := json.Unmarshal(bytes, &buildenv); err != nil {
		fmt.Print(console.SyncFailed(err))
		os.Exit(1)
	}

	if err := buildenv.SyncRepo(buildenv.ConfRepo, buildenv.ConfRepoRef); err != nil {
		fmt.Print(console.SyncFailed(err))
		os.Exit(1)
	}

	return true
}
