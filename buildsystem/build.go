package buildsystem

import (
	"buildenv/generator"
	"buildenv/pkg/color"
	pkgio "buildenv/pkg/io"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type BuildSystem interface {
	Clone(repoUrl, repoRef string) error
	Configure(buildType string) error
	Build() error
	Install() error
}

type BuildConfig struct {
	Pattern     string                 `json:"pattern"`
	BuildTool   string                 `json:"build_tool"`
	Arguments   []string               `json:"arguments"`
	Depedencies []string               `json:"dependencies"`
	CMakeConfig *generator.CMakeConfig `json:"cmake_config"`

	// Internal fields
	Version      string
	SystemName   string
	LibName      string
	SourceDir    string
	SourceFolder string // Some thirdpartys' source code is not in the root folder, so we need to specify it.
	BuildDir     string
	InstalledDir string
	JobNum       int
}

func (b BuildConfig) Verify() error {
	if b.BuildTool == "" {
		return fmt.Errorf("build_tool is empty")
	}

	return nil
}

func (b BuildConfig) Clone(repoUrl, repoRef string) error {
	var commands []string

	// Clone repo or sync repo.
	if pkgio.PathExists(b.SourceDir) {
		// Change to source dir to execute git command.
		if err := os.Chdir(b.SourceDir); err != nil {
			return err
		}

		commands = append(commands, "git reset --hard && git clean -xfd")
		commands = append(commands, fmt.Sprintf("git -C %s fetch origin", b.SourceDir))
		commands = append(commands, fmt.Sprintf("git -C %s checkout %s", b.SourceDir, repoRef))
		commands = append(commands, fmt.Sprintf("git -C %s pull origin %s", b.SourceDir, repoRef))
	} else {
		commands = append(commands, fmt.Sprintf("git clone --branch %s %s %s", repoRef, repoUrl, b.SourceDir))
	}

	// Execute clone command.
	commandLine := strings.Join(commands, " && ")

	if err := b.execute("[buildenv clone]", commandLine, ""); err != nil {
		return err
	}

	return nil
}

func (b BuildConfig) execute(summary, command, logPath string) error {
	fmt.Print(color.Sprintf(color.Blue, "\n%s: %s\n\n", summary, command))

	// Create command for windows and linux.
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("bash", "-c", command)
	}

	// Create log file if log path specified.
	if logPath != "" {
		if err := os.MkdirAll(filepath.Dir(logPath), os.ModeDir|os.ModePerm); err != nil {
			return err
		}
		logFile, err := os.Create(logPath)
		if err != nil {
			return err
		}
		defer logFile.Close()

		// Write command summary as header content of file.
		io.WriteString(logFile, fmt.Sprintf("%s: %s\n\n", summary, command))

		cmd.Stdout = io.MultiWriter(os.Stdout, logFile)
		cmd.Stderr = io.MultiWriter(os.Stderr, logFile)
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
	}

	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (b BuildConfig) CheckAndRepair(url, version, buildType string, cmakeConfig *generator.CMakeConfig) error {
	var buildSystem BuildSystem

	switch b.BuildTool {
	case "cmake":
		buildSystem = NewCMake(b)
	case "ninja":
		buildSystem = NewNinja(b)
	case "make":
		buildSystem = NewMake(b)
	case "autotools":
		buildSystem = NewAutoTool(b)
	case "meson":
		buildSystem = NewMeson(b)
	default:
		return fmt.Errorf("unsupported build system: %s", b.BuildTool)
	}

	if err := buildSystem.Clone(url, version); err != nil {
		return err
	}

	if err := buildSystem.Configure(buildType); err != nil {
		return err
	}

	if err := buildSystem.Build(); err != nil {
		return err
	}

	if err := buildSystem.Install(); err != nil {
		return err
	}

	// Generate cmake config.
	if cmakeConfig != nil {
		cmakeConfig.Version = b.Version
		cmakeConfig.SystemName = b.SystemName
		cmakeConfig.Libname = b.LibName
		cmakeConfig.BuildType = buildType
		if err := cmakeConfig.Generate(b.InstalledDir); err != nil {
			return err
		}
	}
	return nil
}
