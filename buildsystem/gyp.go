package buildsystem

import (
	"buildenv/pkg/cmd"
	"buildenv/pkg/fileio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func NewGyp(config BuildConfig) *gyp {
	return &gyp{BuildConfig: config}
}

type gyp struct {
	BuildConfig
}

func (g gyp) Configure(buildType string) error {
	return nil
}

func (g gyp) Build() error {
	// Remove build dir and create it for configure process.
	if err := os.RemoveAll(g.PortConfig.BuildDir); err != nil {
		return err
	}

	// Create build dir if not exists.
	if err := os.MkdirAll(g.PortConfig.BuildDir, os.ModeDir|os.ModePerm); err != nil {
		return err
	}

	// Some third-party's configure scripts is not exist in the source folder root.
	g.PortConfig.SourceDir = filepath.Join(g.PortConfig.SourceDir, g.PortConfig.SourceFolder)
	if err := os.Chdir(g.PortConfig.SourceDir); err != nil {
		return err
	}

	joinedArgs := strings.Join(g.Arguments, " ")

	// Execute build.
	logPath := g.getLogPath("build")
	title := fmt.Sprintf("[build %s@%s]", g.PortConfig.LibName, g.PortConfig.LibVersion)
	executor := cmd.NewExecutor(title, "./build.sh "+joinedArgs)
	executor.SetLogPath(logPath)
	if err := executor.Execute(); err != nil {
		return err
	}

	return nil
}

func (g gyp) Install() error {
	headerDir := filepath.Join(g.PortConfig.SourceDir, "dist", "public")
	libDir := filepath.Join(g.PortConfig.SourceDir, "dist", "Debug", "lib")
	binDir := filepath.Join(g.PortConfig.SourceDir, "dist", "Debug", "bin")

	if err := fileio.CopyDir(headerDir, filepath.Join(g.PortConfig.PackageDir, "include")); err != nil {
		return fmt.Errorf("failed to install include of %w", err)
	}

	if err := fileio.CopyDir(libDir, filepath.Join(g.PortConfig.PackageDir, "lib")); err != nil {
		return fmt.Errorf("failed to install lib of %w", err)
	}

	if err := fileio.CopyDir(binDir, filepath.Join(g.PortConfig.PackageDir, "bin")); err != nil {
		return fmt.Errorf("failed to install bin of %w", err)
	}

	return nil
}
