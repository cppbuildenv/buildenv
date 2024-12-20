package config

import (
	"buildenv/pkg/io"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Context interface {
	BuildEnvDir() string
	Platform() Platform
	Project() Project
	Toolchain() *Toolchain
	RootFS() *RootFS
	SystemName() string
	BuildType() string
	JobNum() int
}

func NewBuildEnv(buildType string) *buildenv {
	// Set default build type if not specified.
	if strings.TrimSpace(buildType) == "" {
		buildType = "Release"
	}

	return &buildenv{
		configData: configData{
			JobNum: runtime.NumCPU(),
		},
		buildType: buildType,
	}
}

type buildenv struct {
	configData

	// Internal fields.
	platform  Platform
	project   Project
	buildType string
}

type configData struct {
	ConfRepoUrl  string `json:"conf_repo_url"`
	ConfRepoRef  string `json:"conf_repo_ref"`
	PlatformName string `json:"platform_name"`
	ProjectName  string `json:"project_name"`
	JobNum       int    `json:"job_num"`
}

func (b *buildenv) ChangePlatform(platformName string) error {
	buildEnvPath := filepath.Join(Dirs.WorkspaceDir, "buildenv.json")
	if err := b.init(buildEnvPath); err != nil {
		return err
	}

	b.configData.PlatformName = platformName
	bytes, err := json.MarshalIndent(b, "", "    ")
	if err != nil {
		return fmt.Errorf("cannot marshal buildenv conf: %w", err)
	}
	if err := os.WriteFile(buildEnvPath, bytes, os.ModePerm); err != nil {
		return err
	}

	return nil
}

func (b *buildenv) ChangeProject(projectName string) error {
	buildEnvPath := filepath.Join(Dirs.WorkspaceDir, "buildenv.json")
	if err := b.init(buildEnvPath); err != nil {
		return err
	}

	b.configData.ProjectName = projectName
	bytes, err := json.MarshalIndent(b, "", "    ")
	if err != nil {
		return fmt.Errorf("cannot marshal buildenv conf: %w", err)
	}
	if err := os.WriteFile(buildEnvPath, bytes, os.ModePerm); err != nil {
		return err
	}

	return nil
}

func (b *buildenv) Verify(args VerifyArgs) error {
	buildEnvPath := filepath.Join(Dirs.WorkspaceDir, "buildenv.json")
	if err := b.init(buildEnvPath); err != nil {
		return err
	}

	// init and verify platform.
	if err := b.platform.Init(b, b.configData.PlatformName); err != nil {
		return err
	}
	if err := b.platform.Verify(args); err != nil {
		return err
	}

	// init and verify project
	if err := b.project.Init(b, b.configData.ProjectName); err != nil {
		return err
	}
	if err := b.project.Verify(args); err != nil {
		return err
	}

	return nil
}

func (b buildenv) Synchronize(repo, ref string) (string, error) {
	if b.ConfRepoUrl == "" {
		return "", fmt.Errorf("no conf repo has been provided for buildenv")
	}

	if b.ConfRepoRef == "" {
		return "", fmt.Errorf("no conf repo ref has been provided for buildenv")
	}

	var commands []string

	// Clone or git checkout repo.
	confDir := filepath.Join(Dirs.WorkspaceDir, "conf")
	if io.PathExists(confDir) {
		// clean up and checkout to ref.
		if io.PathExists(filepath.Join(confDir, ".git")) {
			// cd `conf`` to execute git command.
			if err := os.Chdir(confDir); err != nil {
				return "", err
			}

			commands = append(commands, "git reset --hard && git clean -xfd")
			commands = append(commands, fmt.Sprintf("git -C %s fetch", confDir))
			commands = append(commands, fmt.Sprintf("git -C %s checkout %s", confDir, ref))
			commands = append(commands, "git pull")
		} else {
			// clean up and clone.
			commands = append(commands, fmt.Sprintf("rm -rf %s", confDir))
			commands = append(commands, fmt.Sprintf("git clone --branch %s --single-branch %s %s", ref, repo, confDir))
		}
	} else {
		commands = append(commands, fmt.Sprintf("git clone --branch %s --single-branch %s %s", ref, repo, confDir))
	}

	commandLine := strings.Join(commands, " && ")
	// Execute clone command.
	output, err := b.execute(commandLine)
	if err != nil {
		return "", err
	}

	return output, nil
}

func (b *buildenv) init(buildEnvPath string) error {
	if !io.PathExists(buildEnvPath) {
		// Create conf directory.
		if err := os.MkdirAll(filepath.Dir(buildEnvPath), os.ModeDir|os.ModePerm); err != nil {
			return err
		}

		b.configData.JobNum = runtime.NumCPU()

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

func (b buildenv) execute(command string) (string, error) {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("bash", "-c", command)
	}

	var buffer bytes.Buffer
	cmd.Stdout = &buffer
	cmd.Stderr = &buffer

	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

// ----------------------- Implementation of BuildEnvContext ----------------------- //

func (b buildenv) BuildEnvDir() string {
	return filepath.Join(Dirs.WorkspaceDir, "conf")
}

func (b buildenv) Platform() Platform {
	return b.platform
}

func (b buildenv) Project() Project {
	return b.project
}

func (b buildenv) Toolchain() *Toolchain {
	return b.platform.Toolchain
}

func (b buildenv) RootFS() *RootFS {
	return b.platform.RootFS
}

func (b buildenv) SystemName() string {
	if b.Toolchain() == nil {
		return runtime.GOOS
	}

	return b.Toolchain().SystemName
}

func (b buildenv) BuildType() string {
	return b.buildType
}

func (b buildenv) JobNum() int {
	return b.configData.JobNum
}
