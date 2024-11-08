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
	var buildenv config.BuildEnv
	if err := buildenv.Write(platformPath); err != nil {
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

	var buildenvConf config.BuildEnvConf
	if err := buildenvConf.Verify(args); err != nil {
		return err
	}

	var buildenv config.BuildEnv
	platformPath := filepath.Join(config.Dirs.PlatformDir, platformName+".json")
	if err := buildenv.Read(platformPath); err != nil {
		return err
	}

	if err := buildenv.Verify(args); err != nil {
		return err
	}

	scriptDir := filepath.Join(config.Dirs.WorkspaceDir, "script")
	if _, err := buildenv.CreateToolchainFile(scriptDir); err != nil {
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
