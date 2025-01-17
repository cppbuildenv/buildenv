package buildsystem

import (
	"buildenv/generator"
	"buildenv/pkg/fileio"
	"fmt"
	"os"
	"path/filepath"
	"slices"
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
	PortsDir     string // ${buildenv}/ports
	SourceDir    string // for example: ${buildenv}/buildtrees/ffmpeg/src
	SourceFolder string // Some thirdpartys' source code is not in the root folder, so we need to specify it.
	BuildDir     string // for example: ${buildenv}/buildtrees/ffmpeg/x86_64-linux-20.04-Release
	PackageDir   string // ${buildenv}/packages/ffmpeg@n3.4.13-x86_64-linux-20.04-Release
	InstalledDir string // for example: ${buildenv}/installed/x86_64-linux-20.04-Release
	JobNum       int    // number of jobs to run in parallel
}

type BuildSystem interface {
	Clone(repoUrl, repoRef string) error
	SourceEnvs() error
	Patch(repoRef string) error
	Configure(buildType string) error
	Build() error
	Install() error
	PackageFiles(packageDir, platformName, projectName, buildType string) ([]string, error)
}

type patch struct {
	Mode   string   `json:"mode"`
	Refers []string `json:"refers"`
}

type BuildConfig struct {
	Pattern     string   `json:"pattern"`
	BuildTool   string   `json:"build_tool"`
	LibraryType string   `json:"library_type"`
	EnvVars     []string `json:"env_vars"`
	Patches     *patch   `json:"patches"`
	Arguments   []string `json:"arguments"`
	Depedencies []string `json:"dependencies"`
	CMakeConfig string   `json:"cmake_config"`

	// Internal fields
	buildSystem BuildSystem
	PortConfig  PortConfig
}

func (b BuildConfig) Verify() error {
	if b.BuildTool == "" {
		return fmt.Errorf("build_tool is empty, it should be one of cmake, ninja, makefiles, autotools, meson, b2")
	}

	if !slices.Contains([]string{"cmake", "ninja", "makefiles", "autotools", "meson", "b2"}, b.BuildTool) {
		return fmt.Errorf("unsupported build tool: %s, it should be one of cmake, ninja, makefiles, autotools, meson, b2",
			b.BuildTool)
	}

	return nil
}

func (b BuildConfig) Clone(repoUrl, repoRef string) error {
	// Clone repo only when source dir not exists.
	if !fileio.PathExists(b.PortConfig.SourceDir) {
		var commands []string
		commands = append(commands, fmt.Sprintf("git clone --branch %s %s %s", repoRef, repoUrl, b.PortConfig.SourceDir))

		// Execute clone command.
		commandLine := strings.Join(commands, " && ")
		title := fmt.Sprintf("[clone %s]", b.PortConfig.LibName)
		if err := execute(title, commandLine, ""); err != nil {
			return err
		}
	}

	return nil
}

func (b BuildConfig) Patch(repoRef string) error {
	if b.Patches == nil || len(b.Patches.Refers) == 0 {
		return nil
	}

	switch b.Patches.Mode {
	case "cherry-pick":
		title := fmt.Sprintf("[patch %s]", b.PortConfig.LibName)
		if err := cherryPick(title, b.PortConfig.SourceDir, b.Patches.Refers); err != nil {
			return err
		}

	case "rebase":
		title := fmt.Sprintf("[patch %s]", b.PortConfig.LibName)
		if err := rebase(title, b.PortConfig.SourceDir, repoRef, b.Patches.Refers); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported patch mode: %s", b.Patches.Mode)
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

func (b *BuildConfig) Install(url, version, buildType string) error {
	if err := b.buildSystem.Clone(url, version); err != nil {
		return err
	}
	if err := b.buildSystem.SourceEnvs(); err != nil {
		return err
	}
	if err := b.buildSystem.Patch(version); err != nil {
		return err
	}
	if err := b.buildSystem.Configure(buildType); err != nil {
		return err
	}
	if err := b.buildSystem.Build(); err != nil {
		return err
	}
	if err := b.buildSystem.Install(); err != nil {
		return err
	}

	// Some pkg-config file may have absolute path,
	// so we need to replace them with relative path.
	if err := fixupPkgConfig(b.PortConfig.PackageDir); err != nil {
		return fmt.Errorf("fixup pkg-config failed: %w", err)
	}

	// Generate cmake config.
	portDir := filepath.Join(b.PortConfig.PortsDir, b.PortConfig.LibName)
	cmakeConfig, err := generator.FindMatchedConfig(portDir, b.PortConfig.LibVersion, b.CMakeConfig)
	if err != nil {
		return err
	}
	if cmakeConfig != nil {
		cmakeConfig.Version = b.PortConfig.LibVersion
		cmakeConfig.SystemName = b.PortConfig.SystemName
		cmakeConfig.Libname = b.PortConfig.LibName
		cmakeConfig.BuildType = buildType
		if err := cmakeConfig.Generate(b.PortConfig.PackageDir); err != nil {
			return err
		}
	}
	return nil
}

func (b *BuildConfig) InitBuildSystem() error {
	switch b.BuildTool {
	case "cmake":
		b.buildSystem = NewCMake(*b)
	case "ninja":
		b.buildSystem = NewNinja(*b)
	case "makefiles":
		b.buildSystem = NewMakefiles(*b)
	case "autotools":
		b.buildSystem = NewAutoTool(*b)
	case "meson":
		b.buildSystem = NewMeson(*b)
	case "b2":
		b.buildSystem = NewB2(*b)
	default:
		return fmt.Errorf("unsupported build system: %s", b.BuildTool)
	}

	return nil
}

func (b BuildConfig) PackageFiles(packageDir, platformName, projectName, buildType string) ([]string, error) {
	if !fileio.PathExists(packageDir) {
		return nil, nil
	}

	var files []string
	if err := filepath.WalkDir(packageDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		relativePath, err := filepath.Rel(packageDir, path)
		if err != nil {
			return err
		}

		platformProject := fmt.Sprintf("%s@%s@%s", platformName, projectName, buildType)
		files = append(files, platformProject+"/"+relativePath)
		return nil
	}); err != nil {
		return nil, err
	}

	return files, nil
}

func (b BuildConfig) BuildSystem() BuildSystem {
	return b.buildSystem
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
