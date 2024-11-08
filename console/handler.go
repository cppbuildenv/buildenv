package console

import (
	"buildenv/config"
	"fmt"
	"os"
	"path/filepath"
)

var PlatformCallbacks = platformCallbacks{}

type platformCallbacks struct{}

func (p platformCallbacks) OnCreatePlatform(platformName string) error {
	if platformName == "" {
		return fmt.Errorf("platform name is empty")
	}

	// Check if same platform exists.
	platformPath := filepath.Join(config.Dirs.PlatformDir, platformName+".json")
	if pathExists(platformPath) {
		return fmt.Errorf("[%s] already exists", platformPath)
	}

	// Create platform file.
	var platform config.Platform
	if err := platform.Write(platformPath); err != nil {
		return err
	}

	return nil
}

func (p platformCallbacks) OnSelectPlatform(platformName string) error {
	args := config.VerifyArgs{
		Silent:         false,
		CheckAndRepair: false,
		BuildType:      "Release",
	}

	var buildenvConf config.BuildEnv
	if err := buildenvConf.Verify(args); err != nil {
		return err
	}

	var platform config.Platform
	platformPath := filepath.Join(config.Dirs.PlatformDir, platformName+".json")
	if err := platform.Read(platformPath); err != nil {
		return err
	}

	if err := platform.Verify(args); err != nil {
		return err
	}

	scriptDir := filepath.Join(config.Dirs.WorkspaceDir, "script")
	if _, err := platform.CreateToolchainFile(scriptDir); err != nil {
		return err
	}

	return nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	return !os.IsNotExist(err)
}
