package buildsystem

import (
	"buildenv/pkg/cmd"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func NewGpy(config BuildConfig) *gpy {
	return &gpy{BuildConfig: config}
}

type gpy struct {
	BuildConfig
}

func (c gpy) Configure(buildType string) error {
	return nil
}

func (c gpy) Build() error {
	// Some third-party's configure scripts is not exist in the source folder root.
	c.PortConfig.SourceDir = filepath.Join(c.PortConfig.SourceDir, c.PortConfig.SourceFolder)
	if err := os.Chdir(c.PortConfig.SourceDir); err != nil {
		return err
	}

	joinedArgs := strings.Join(c.Arguments, " ")

	// Execute build.
	logPath := c.getLogPath("build")
	title := fmt.Sprintf("[build %s@%s]", c.PortConfig.LibName, c.PortConfig.LibVersion)
	executor := cmd.NewExecutor(title, "./build.sh "+joinedArgs)
	executor.SetLogPath(logPath)
	if err := executor.Execute(); err != nil {
		return err
	}

	return nil
}

func (c gpy) Install() error {
	return nil
}
