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
	Configure(buildType string) (string, error)
	Build() (string, error)
	Install() (string, error)
	InstalledFiles(installLogFile string) ([]string, error)
}

type BuildConfig struct {
	Pattern     string                 `json:"pattern"`
	BuildTool   string                 `json:"build_tool"`
	Arguments   []string               `json:"arguments"`
	Depedencies []string               `json:"dependencies"`
	CMakeConfig *generator.CMakeConfig `json:"cmake_config"`

	// Internal fields
	BuildSystem      BuildSystem
	Version          string
	SystemName       string
	LibName          string
	SourceDir        string
	SourceFolder     string // Some thirdpartys' source code is not in the root folder, so we need to specify it.
	BuildDir         string
	InstalledDir     string
	InstalledRootDir string
	JobNum           int
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
	title := fmt.Sprintf("[clone %s]", b.LibName)
	if err := b.execute(title, commandLine, ""); err != nil {
		return err
	}

	return nil
}

func (b BuildConfig) execute(title, command, logPath string) error {
	fmt.Print(color.Sprintf(color.Blue, "\n%s: %s\n\n", title, command))

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
		io.WriteString(logFile, fmt.Sprintf("%s: %s\n\n", title, command))

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

func (b *BuildConfig) CheckAndRepair(url, version, buildType string, cmakeConfig *generator.CMakeConfig) (string, error) {
	switch b.BuildTool {
	case "cmake":
		b.BuildSystem = NewCMake(*b)
	case "ninja":
		b.BuildSystem = NewNinja(*b)
	case "make":
		b.BuildSystem = NewMake(*b)
	case "autotools":
		b.BuildSystem = NewAutoTool(*b)
	case "meson":
		b.BuildSystem = NewMeson(*b)
	default:
		return "", fmt.Errorf("unsupported build system: %s", b.BuildTool)
	}

	if err := b.BuildSystem.Clone(url, version); err != nil {
		return "", err
	}

	if _, err := b.BuildSystem.Configure(buildType); err != nil {
		return "", err
	}

	if _, err := b.BuildSystem.Build(); err != nil {
		return "", err
	}

	installLogPath, err := b.BuildSystem.Install()
	if err != nil {
		return "", err
	}

	// Generate cmake config.
	if cmakeConfig != nil {
		cmakeConfig.Version = b.Version
		cmakeConfig.SystemName = b.SystemName
		cmakeConfig.Libname = b.LibName
		cmakeConfig.BuildType = buildType
		if err := cmakeConfig.Generate(b.InstalledDir); err != nil {
			return "", err
		}
	}
	return installLogPath, nil
}
