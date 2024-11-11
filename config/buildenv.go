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

	InstalledDir string `json:"-"`
}

func (b *BuildEnv) ChangePlatform(platform string) error {
	buildEnvPath := filepath.Join(Dirs.WorkspaceDir, "conf", "buildenv.json")
	if err := b.init(buildEnvPath, "Release"); err != nil {
		return err
	}

	b.Platform = platform
	bytes, err := json.MarshalIndent(b, "", "    ")
	if err != nil {
		return fmt.Errorf("cannot marshal buildenv conf: %w", err)
	}
	if err := os.WriteFile(buildEnvPath, bytes, os.ModePerm); err != nil {
		return err
	}

	return nil
}

func (b *BuildEnv) Verify(args VerifyArgs) error {
	buildEnvPath := filepath.Join(Dirs.WorkspaceDir, "conf", "buildenv.json")
	if err := b.init(buildEnvPath, args.BuildType); err != nil {
		return err
	}
	if b.Platform == "" {
		return fmt.Errorf("no platform has been selected for buildenv")
	}

	// Check if platform file exists and read it.
	platformPath := filepath.Join(Dirs.PlatformDir, b.Platform+".json")
	if !pathExists(platformPath) {
		return fmt.Errorf("platform file not exists: %s", platformPath)
	}

	var platform Platform
	if err := platform.Init(platformPath, b.InstalledDir); err != nil {
		return err
	}

	// Verify buildenv, it'll verify toolchain, tools and dependencies inside.
	if err := platform.Verify(args); err != nil {
		return err
	}

	return nil
}

func (b *BuildEnv) init(buildEnvPath, buildType string) error {
	if !pathExists(buildEnvPath) {
		// Create conf directory.
		if err := os.MkdirAll(filepath.Dir(buildEnvPath), os.ModeDir|os.ModePerm); err != nil {
			return err
		}

		b.JobNum = runtime.NumCPU()

		// Create buildenv conf file with default values.
		bytes, err := json.MarshalIndent(b, "", "    ")
		if err != nil {
			return fmt.Errorf("cannot marshal buildenv conf: %w", err)
		}
		if err := os.WriteFile(buildEnvPath, bytes, os.ModePerm); err != nil {
			return err
		}

		return nil
	}

	// Rewrite buildenv file with new platform.
	bytes, err := os.ReadFile(buildEnvPath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bytes, b); err != nil {
		return err
	}

	// Set values of internal fields.
	b.InstalledDir = filepath.Join(Dirs.InstalledRootDir, b.Platform+"-"+buildType)

	return nil
}
