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
	"unicode"
)

type PortConfig struct {
	SystemName      string // like: `Linux`, `Darwin`, `Windows`
	SystemProcessor string // like: `aarch64`, `x86_64`, `i386`
	Host            string // like: `aarch64-linux-gnu`
	RootFS          string // absolute path of rootfs
	ToolchainPrefix string // like: `aarch64-linux-gnu-`
	LibName         string // like: `ffmpeg`
	LibVersion      string // like: `4.4`

	// Internal fields
	PortsDir         string // absolute path of ports dir
	SourceDir        string // absolute path of source code
	SourceFolder     string // Some thirdpartys' source code is not in the root folder, so we need to specify it.
	BuildDir         string // absolute path of build dir
	InstalledDir     string // absolute path of installed dir
	InstalledRootDir string // absolute path of installed root dir
	JobNum           int    // number of jobs to run in parallel
}

type BuildSystem interface {
	Clone(repoUrl, repoRef string) error
	SourceEnvs() error
	Patch(repoRef string) error
	Configure(buildType string) (string, error)
	Build() (string, error)
	Install() (string, error)
	InstalledFiles(installLogFile string) ([]string, error)
}

type patch struct {
	Mode string `json:"mode"`
	Ref  string `json:"ref"`
}

type BuildConfig struct {
	Pattern     string   `json:"pattern"`
	BuildTool   string   `json:"build_tool"`
	EnvVars     []string `json:"env_vars"`
	Patches     []patch  `json:"patches"`
	Arguments   []string `json:"arguments"`
	Depedencies []string `json:"dependencies"`
	CMakeConfig string   `json:"cmake_config"`

	// Internal fields
	buildSystem BuildSystem
	portConfig  PortConfig
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
	if pkgio.PathExists(b.portConfig.SourceDir) {
		// Change to source dir to execute git command.
		if err := os.Chdir(b.portConfig.SourceDir); err != nil {
			return err
		}

		commands = append(commands, "git reset --hard && git clean -xfd")
		commands = append(commands, fmt.Sprintf("git -C %s fetch origin", b.portConfig.SourceDir))
		commands = append(commands, fmt.Sprintf("git -C %s checkout %s", b.portConfig.SourceDir, repoRef))
		commands = append(commands, fmt.Sprintf("git -C %s pull origin %s", b.portConfig.SourceDir, repoRef))
	} else {
		commands = append(commands, fmt.Sprintf("git clone --branch %s %s %s", repoRef, repoUrl, b.portConfig.SourceDir))
	}

	// Execute clone command.
	commandLine := strings.Join(commands, " && ")
	title := fmt.Sprintf("[clone %s]", b.portConfig.LibName)
	if err := b.execute(title, commandLine, ""); err != nil {
		return err
	}

	return nil
}

func (b BuildConfig) Patch(repoRef string) error {
	if len(b.Patches) == 0 {
		return nil
	}

	// Change to source dir to execute git command.
	if err := os.Chdir(b.portConfig.SourceDir); err != nil {
		return err
	}

	// Execute patch command.
	var commands []string
	commands = append(commands, "git reset --hard && git clean -xfd")
	commands = append(commands, fmt.Sprintf("git -C %s fetch origin", b.portConfig.SourceDir))

	for _, patch := range b.Patches {
		switch patch.Mode {
		case "cherry-pick":
			commands = append(commands, fmt.Sprintf("git cherry-pick %s", patch.Ref))

		case "rebase":
			commands = append(commands, fmt.Sprintf("git checkout %s", patch.Ref))
			commands = append(commands, fmt.Sprintf("git rebase %s", repoRef))

		default:
			return fmt.Errorf("unsupported patch mode: %s", patch.Mode)
		}
	}

	commandLine := strings.Join(commands, " && ")
	title := fmt.Sprintf("[patch %s]", b.portConfig.LibName)
	if err := b.execute(title, commandLine, ""); err != nil {
		return err
	}

	return nil
}

func (b BuildConfig) SourceEnvs() error {
	for _, item := range b.EnvVars {
		item = strings.TrimSpace(item)

		index := strings.Index(item, "=")
		if index == -1 {
			return fmt.Errorf("invalid env var: %s", item)
		}

		key := strings.TrimSpace(item[:index])
		value := strings.TrimSpace(item[index+1:])

		if err := b.validateEnv(key); err != nil {
			return err
		}

		os.Setenv(key, value)
	}

	return nil
}

func (b *BuildConfig) Install(url, version, buildType string) (string, error) {
	switch b.BuildTool {
	case "cmake":
		b.buildSystem = NewCMake(*b)
	case "ninja":
		b.buildSystem = NewNinja(*b)
	case "make":
		b.buildSystem = NewMake(*b)
	case "autotools":
		b.buildSystem = NewAutoTool(*b)
	case "meson":
		b.buildSystem = NewMeson(*b)
	default:
		return "", fmt.Errorf("unsupported build system: %s", b.BuildTool)
	}

	if err := b.buildSystem.Clone(url, version); err != nil {
		return "", err
	}
	if err := b.buildSystem.SourceEnvs(); err != nil {
		return "", err
	}
	if err := b.buildSystem.Patch(version); err != nil {
		return "", err
	}
	if _, err := b.buildSystem.Configure(buildType); err != nil {
		return "", err
	}
	if _, err := b.buildSystem.Build(); err != nil {
		return "", err
	}
	installLogPath, err := b.buildSystem.Install()
	if err != nil {
		return "", err
	}

	// Generate cmake config.
	portDir := filepath.Join(b.portConfig.PortsDir, b.portConfig.LibName)
	cmakeConfig, err := generator.FindMatchedConfig(portDir, b.CMakeConfig)
	if err != nil {
		return "", err
	}
	if cmakeConfig != nil {
		cmakeConfig.Version = b.portConfig.LibVersion
		cmakeConfig.SystemName = b.portConfig.SystemName
		cmakeConfig.Libname = b.portConfig.LibName
		cmakeConfig.BuildType = buildType
		if err := cmakeConfig.Generate(b.portConfig.InstalledDir); err != nil {
			return "", err
		}
	}
	return installLogPath, nil
}

func (b BuildConfig) BuildSystem() BuildSystem {
	return b.buildSystem
}

func (b *BuildConfig) SetPortConfig(portConfig PortConfig) {
	b.portConfig = portConfig
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

func (b BuildConfig) validateEnv(envVar string) error {
	envVar = strings.TrimSpace(envVar)
	parts := strings.Split(envVar, "=")
	if len(parts) == 1 {
		if strings.Contains(envVar, " ") ||
			strings.Contains(envVar, "-") ||
			strings.Contains(envVar, "&") ||
			strings.Contains(envVar, "!") ||
			strings.Contains(envVar, "\\") ||
			strings.Contains(envVar, "|") ||
			strings.Contains(envVar, ";") ||
			strings.Contains(envVar, "'") ||
			strings.Contains(envVar, "#") ||
			unicode.IsDigit(rune(envVar[0])) {
			return fmt.Errorf("invalid env key: %s", envVar)
		}
	}
	return nil
}
