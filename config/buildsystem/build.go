package buildsystem

import (
	"buildenv/config/generator"
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
	Clone(repo, ref string) error
	Configure(buildType string) error
	Build() error
	Install() error
}

type BuildConfig struct {
	Pattern             string                     `json:"pattern"`
	BuildTool           string                     `json:"build_tool"`
	Arguments           []string                   `json:"arguments"`
	GenerateCMakeConfig *generator.GeneratorConfig `json:"generate_cmake_config"`

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

func (b BuildConfig) Clone(repo, ref string) error {
	var commands []string

	// Clone repo or sync repo.
	if pkgio.PathExists(b.SourceDir) {
		// Change to source dir to execute git command.
		if err := os.Chdir(b.SourceDir); err != nil {
			return err
		}

		commands = append(commands, "git reset --hard && git clean -xfd")
		commands = append(commands, fmt.Sprintf("git -C %s fetch", b.SourceDir))
		commands = append(commands, fmt.Sprintf("git -C %s checkout %s", b.SourceDir, ref))
		commands = append(commands, "git pull")
	} else {
		commands = append(commands, fmt.Sprintf("git clone --branch %s --single-branch %s %s", ref, repo, b.SourceDir))
	}

	// Assemble cloneLogPath.
	cloneLogPath := filepath.Join(filepath.Dir(b.BuildDir), filepath.Base(b.BuildDir)+"-clone.log")

	// Execute clone command.
	commandLine := strings.Join(commands, " && ")
	if err := b.execute(commandLine, cloneLogPath); err != nil {
		return err
	}

	return nil
}

func (b BuildConfig) execute(command, logPath string) error {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("bash", "-c", command)
	}

	// Create log file if not exsit.
	if err := os.MkdirAll(filepath.Dir(logPath), os.ModeDir|os.ModePerm); err != nil {
		return err
	}
	logFile, err := os.Create(logPath)
	if err != nil {
		return err
	}
	defer logFile.Close()

	outWriter := color.NewWriter(os.Stdout, color.Green)
	cmd.Stdout = io.MultiWriter(outWriter, logFile)

	errWriter := color.NewWriter(os.Stderr, color.Red)
	cmd.Stderr = io.MultiWriter(errWriter, logFile)

	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (b BuildConfig) CheckAndRepair(url, version, buildType string) error {
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

	if b.GenerateCMakeConfig != nil {
		b.GenerateCMakeConfig.Version = b.Version
		b.GenerateCMakeConfig.SystemName = b.SystemName
		b.GenerateCMakeConfig.Libname = b.LibName
		b.GenerateCMakeConfig.BuildType = buildType
		if err := b.GenerateCMakeConfig.Generate(b.InstalledDir); err != nil {
			return err
		}
	}
	return nil
}
