package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func newSelectPlatformCmd(platformDir string, selected func(fullpath string)) *selectPlatformCmd {
	return &selectPlatformCmd{
		platformDir: platformDir,
		selected:    selected,
	}
}

type selectPlatformCmd struct {
	platformName string
	platformDir  string
	selected     func(fullpath string)
}

func (p *selectPlatformCmd) register() {
	flag.StringVar(&p.platformName, "sp", "", "select a platform as build target platform")
	flag.StringVar(&p.platformName, "select_platform", "", "select a platform as build target platform")
}

func (s *selectPlatformCmd) listen() (handled bool) {
	if s.platformName == "" {
		return false
	}

	if strings.HasSuffix(s.platformName, ".json") {
		s.platformName = filepath.Join(s.platformName, ".json")

	}
	filePath := filepath.Join(s.platformDir, s.platformName)
	if s.pathExists(filePath) {
		fmt.Printf("[✘] ---- none exist platform: %s\n", s.platformName)
		return true
	}

	// TODO: validate platform json file

	s.selected(filePath)
	fmt.Printf("[✔] ---- build target platform: %s\n", s.platformName)

	return true
}

func (selectPlatformCmd) pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	return !os.IsNotExist(err)
}
