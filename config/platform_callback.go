package config

import (
	"fmt"
	"path/filepath"
)

var PlatformCallbacksImpl = platformCallbacksImpl{}

type platformCallbacksImpl struct{}

func (p platformCallbacksImpl) OnCreatePlatform(platformPath string) error {
	if platformPath == "" {
		return fmt.Errorf("platformPath is empty")
	}

	// Create platform file.
	var platform Platform
	if err := platform.Write(platformPath); err != nil {
		return err
	}

	return nil
}

func (p platformCallbacksImpl) OnSelectPlatform(platformName string) error {
	buildenv := NewBuildEnv("Release")
	if err := buildenv.Verify(NewVerifyArgs(false, false, "Release")); err != nil {
		return err
	}

	if err := buildenv.ChangePlatform(platformName); err != nil {
		return err
	}

	var platform Platform
	if err := platform.Init(buildenv, platformName); err != nil {
		return err
	}

	args := NewVerifyArgs(false, false, "Release")
	if err := platform.Verify(args); err != nil {
		return err
	}

	scriptDir := filepath.Join(Dirs.WorkspaceDir, "script")
	if _, err := platform.CreateToolchainFile(scriptDir); err != nil {
		return err
	}

	return nil
}
