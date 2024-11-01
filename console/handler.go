package console

import (
	"buildenv/config"
	"fmt"
	"os"
)

var PlatformCallbacks = platformCallbacks{}

type platformCallbacks struct{}

func (p platformCallbacks) OnCreatePlatform(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("platform name is empty")
	}

	// Check if same platform exists.
	if pathExists(filePath) {
		return fmt.Errorf("[%s] already exists", filePath)
	}

	// Create platform file.
	var buildenv config.BuildEnv
	if err := buildenv.Write(filePath); err != nil {
		return err
	}

	return nil
}

func (p platformCallbacks) OnSelectPlatform(filePath string) error {
	var buildenv config.BuildEnv
	if err := buildenv.Read(filePath); err != nil {
		return err
	}

	if err := buildenv.Verify(); err != nil {
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
