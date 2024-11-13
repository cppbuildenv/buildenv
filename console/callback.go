package console

import (
	"buildenv/config"
	"fmt"
	"path/filepath"
)

var PlatformCallbacks = platformCallbacks{}

type platformCallbacks struct{}

func (p platformCallbacks) OnCreatePlatform(platformName string) error {
	if platformName == "" {
		return fmt.Errorf("platform name is empty")
	}

	platformPath := filepath.Join(config.Dirs.PlatformDir, platformName+".json")

	// Create platform file.
	var platform config.Platform
	if err := platform.Write(platformPath); err != nil {
		return err
	}

	return nil
}

func (p platformCallbacks) OnSelectPlatform(platformName string) error {
	var buildenv config.BuildEnv
	if err := buildenv.ChangePlatform(platformName); err != nil {
		return err
	}

	var platform config.Platform
	platformPath := filepath.Join(config.Dirs.PlatformDir, platformName+".json")
	if err := platform.Init(platformPath); err != nil {
		return err
	}

	args := config.VerifyArgs{
		Silent:         false,
		CheckAndRepair: false,
		BuildType:      "Release",
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
