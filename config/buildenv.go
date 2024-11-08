package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type BuildEnv struct {
	Platform string `json:"platform"`
	ConfRepo string `json:"conf_repo"`
	JobNum   int    `json:"job_num"`
}

func (b *BuildEnv) Verify(args VerifyArgs) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot get current directory: %w", err)
	}

	// Check if buildenv conf file exists.
	buildEnvConfPath := filepath.Join(currentDir, "conf/buildenv.json")
	if !pathExists(buildEnvConfPath) {
		// Create conf directory.
		if err := os.MkdirAll(filepath.Dir(buildEnvConfPath), os.ModeDir|os.ModePerm); err != nil {
			return err
		}

		// Set max job num.
		b.JobNum = runtime.NumCPU()

		// Create buildenv conf file with default values.
		bytes, err := json.MarshalIndent(b, "", "    ")
		if err != nil {
			return fmt.Errorf("cannot marshal buildenv conf: %w", err)
		}
		if err := os.WriteFile(buildEnvConfPath, bytes, os.ModePerm); err != nil {
			return err
		}

		return fmt.Errorf("no platform has been selected for buildenv")
	}

	bytes, err := os.ReadFile(filepath.Join(currentDir, "conf/buildenv.json"))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bytes, b); err != nil {
		return err
	}

	if b.Platform == "" {
		return fmt.Errorf("no platform has been selected for buildenv")
	}

	platformPath := filepath.Join(Dirs.PlatformDir, b.Platform)
	if !pathExists(platformPath) {
		return fmt.Errorf("platform file not exists: %s", platformPath)
	}

	var buildenv Platform
	if err := buildenv.Read(platformPath); err != nil {
		return err
	}

	if err := buildenv.Verify(args); err != nil {
		return err
	}

	return nil
}
