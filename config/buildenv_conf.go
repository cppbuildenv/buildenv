package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type BuildEnvConf struct {
	Platform string `json:"platform"`
	ConfRepo string `json:"conf_repo"`
}

func (b *BuildEnvConf) Verify() error {
	bytes, err := os.ReadFile("conf/buildenv.json")
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, b); err != nil {
		return err
	}

	if b.Platform == "" {
		return fmt.Errorf("platform is empty")
	}

	filePath := filepath.Join(PlatformDir, b.Platform)
	if !pathExists(filePath) {
		return fmt.Errorf("platform file not exists: %s", filePath)
	}

	var buildenv BuildEnv
	if err := buildenv.Read(filePath); err != nil {
		return err
	}

	if err := buildenv.Verify(false); err != nil {
		return err
	}

	return nil
}
