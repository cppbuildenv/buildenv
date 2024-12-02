package config

import (
	"fmt"
	"path/filepath"
)

var Callbacks = callbackImpl{}

type callbackImpl struct{}

func (c callbackImpl) OnCreatePlatform(platformName string) error {
	if platformName == "" {
		return fmt.Errorf("platformName is empty for creating new platform")
	}

	// Create platform file.
	platformPath := filepath.Join(Dirs.PlatformDir, platformName+".json")
	var platform Platform
	if err := platform.Write(platformPath); err != nil {
		return err
	}

	return nil
}

func (c callbackImpl) OnSelectPlatform(platformName string) error {
	// In config mode, we always regard build type as `Release`.
	buildType := "Release"

	buildenv := NewBuildEnv(buildType)
	if err := buildenv.ChangePlatform(platformName); err != nil {
		return err
	}

	// Verify buildenv to check if all the required fields are set for generate toolchain file.
	if err := buildenv.Verify(NewVerifyArgs(false, false, buildType)); err != nil {
		return err
	}

	// Generate toolchain file.
	scriptDir := filepath.Join(Dirs.WorkspaceDir, "script")
	if _, err := buildenv.platform.GenerateToolchainFile(scriptDir); err != nil {
		return err
	}

	return nil
}

func (p callbackImpl) SetOffline(offline bool) error {
	// In config mode, we always regard build type as `Release`.
	buildType := "Release"
	buildenv := NewBuildEnv(buildType)
	return buildenv.SetOffline(offline)
}
