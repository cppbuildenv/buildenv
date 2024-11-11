package build

import (
	"buildenv/pkg/color"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type BuildSystem interface {
	Clone(repo, ref string) error
	Configure(buildType string) error
	Build() error
	Install() error
}

type BuildConfig struct {
	BuildTool   string   `json:"build_tool"`
	Arguments   []string `json:"arguments"`
	Depedencies []string `json:"dependencies"`

	// Internal fields
	SourceDir    string `json:"-"`
	BuildDir     string `json:"-"`
	JobNum       int    `json:"-"`
	InstalledDir string `json:"-"`
}

func (b BuildConfig) Clone(repo, ref string) error {
	var commands []string

	// Clone repo or sync repo.
	if pathExists(b.SourceDir) {
		commands = append(commands, fmt.Sprintf("git -C %s fetch", b.SourceDir))
		commands = append(commands, fmt.Sprintf("git -C %s checkout %s", b.SourceDir, ref))
	} else {
		commands = append(commands, fmt.Sprintf("git clone --branch %s --single-branch %s %s", ref, repo, b.SourceDir))
	}

	// Assemble cloneLogPath.
	cloneLogPath := filepath.Join(filepath.Dir(b.BuildDir), filepath.Base(b.BuildDir)+"-clone.log")

	// Execute clone command.
	for _, command := range commands {
		if err := b.execute(command, cloneLogPath); err != nil {
			return err
		}
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

	outWriter := color.NewWriter(os.Stdout, color.Blue)
	cmd.Stdout = io.MultiWriter(outWriter, logFile)

	errWriter := color.NewWriter(os.Stdout, color.Red)
	cmd.Stderr = io.MultiWriter(errWriter, logFile)

	if err := cmd.Run(); err != nil {
		color.Println(color.Red, fmt.Sprintf("Error execute command: %s", err.Error()))
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
	return nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	return !os.IsNotExist(err)
}