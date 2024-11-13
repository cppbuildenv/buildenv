package config

import (
	"buildenv/pkg/color"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type BuildEnv struct {
	Platform    string `json:"platform"`
	ConfRepo    string `json:"conf_repo"`
	ConfRepoRef string `json:"conf_repo_ref"`
	JobNum      int    `json:"job_num"`
}

func (b *BuildEnv) ChangePlatform(platform string) error {
	buildEnvPath := filepath.Join(Dirs.WorkspaceDir, "buildenv.json")
	if err := b.init(buildEnvPath); err != nil {
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
	buildEnvPath := filepath.Join(Dirs.WorkspaceDir, "buildenv.json")
	if err := b.init(buildEnvPath); err != nil {
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
	if err := platform.Init(platformPath); err != nil {
		return err
	}

	// Verify buildenv, it'll verify toolchain, tools and dependencies inside.
	if err := platform.Verify(args); err != nil {
		return err
	}

	return nil
}

func (b BuildEnv) SyncRepo(repo, ref string) error {
	if b.ConfRepo == "" {
		return fmt.Errorf("no conf repo has been provided for buildenv")
	}

	if b.ConfRepoRef == "" {
		return fmt.Errorf("no conf repo ref has been provided for buildenv")
	}

	var commands []string

	sourceDir := filepath.Join(Dirs.WorkspaceDir, "conf")

	// Clone repo or sync repo.
	if pathExists(sourceDir) {
		commands = append(commands, fmt.Sprintf("git -C %s fetch", sourceDir))
		commands = append(commands, fmt.Sprintf("git -C %s checkout %s", sourceDir, ref))
	} else {
		commands = append(commands, fmt.Sprintf("git clone --branch %s --single-branch %s %s", ref, repo, sourceDir))
	}

	// Execute clone command.
	for _, command := range commands {
		if err := b.execute(command); err != nil {
			return err
		}
	}

	return nil
}

func (b *BuildEnv) init(buildEnvPath string) error {
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

	return nil
}

func (b BuildEnv) execute(command string) error {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("bash", "-c", command)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	if err := cmd.Run(); err != nil {
		color.Println(color.Red, fmt.Sprintf("Error execute command: %s", err.Error()))
		return err
	}

	return nil
}
