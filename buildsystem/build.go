package buildsystem

import (
	"buildenv/generator"
	"buildenv/pkg/cmd"
	"buildenv/pkg/env"
	"buildenv/pkg/fileio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

const supportedString = "b2, bazel, cmake, gyp, makefiles, meson, ninja"

var (
	supportedArray = []string{"b2", "bazel", "cmake", "gyp", "makefiles", "meson", "ninja"}
)

type PortConfig struct {
	LibName    string // like: `ffmpeg`
	LibVersion string // like: `4.4`

	// Internal fields
	CrossTools      CrossTools // cross tools like CC, CXX, FC, RANLIB, AR, LD, NM, OBJDUMP, STRIP
	WorkspaceDir    string     // It is the root directory of buildenv workspace.
	PortsDir        string     // ${buildenv}/ports
	DownloadedDir   string     // ${buildenv}/downloads
	SourceDir       string     // for example: ${buildenv}/buildtrees/ffmpeg/src
	SourceFolder    string     // Some thirdpartys' source code is not in the root folder, so we need to specify it.
	BuildDir        string     // for example: ${buildenv}/buildtrees/ffmpeg/x86_64-linux-20.04-Release
	PackageDir      string     // for example: ${buildenv}/packages/ffmpeg-3.4.13-x86_64-linux-20.04-Release
	InstalledDir    string     // for example: ${buildenv}/installed/x86_64-linux-20.04-Release
	InstalledFolder string     // for example: aarch64-linux-gnu-gcc-9.2^project_01_standard^Release
	ExtraHeaderDirs []string   // headers not in standard include path.
	ExtraLibDirs    []string   // libs not in standard lib path.
	JobNum          int        // number of jobs to run in parallel
	TmpDir          string     // for example: ${buildenv}/downloaded/tmp
}

type BuildSystem interface {
	Clone(repoUrl, repoRef string) error
	Patch() error
	Configure(buildType string) error
	Build() error
	Install() error

	fixConfigure() error
	fixBuild() error // Some thirdpartys need extra steps to fix build, for example: nspr.
	appendBuildEnvs() error
	removeBuildEnvs() error
	fillPlaceHolders()
	setBuildType(buildType string)
	getLogPath(suffix string) string
}

type FixWork struct {
	Scripts []string `json:"scripts"`
	WorkDir string   `json:"work_dir"`
}

type BuildConfig struct {
	Pattern        string   `json:"pattern"`
	BuildTool      string   `json:"build_tool"`
	SystemTools    []string `json:"system_tools"`
	LibraryType    string   `json:"library_type"`
	EnvVars        []string `json:"env_vars"`
	FixConfigure   FixWork  `json:"fix_configure"`
	FixBuild       FixWork  `json:"fix_build"`
	Patches        []string `json:"patches"`
	Options        []string `json:"options"`
	Depedencies    []string `json:"dependencies"`
	DevDepedencies []string `json:"dev_dependencies"`
	CMakeConfig    string   `json:"cmake_config"`

	// Internal fields
	AsDev       bool            `json:"-"`
	PortConfig  PortConfig      `json:"-"`
	buildSystem BuildSystem     `json:"-"`
	environment env.Environment `json:"-"`
}

func (b BuildConfig) Validate() error {
	if b.BuildTool == "" {
		return fmt.Errorf("build_tool is empty, it should be one of %s", supportedString)
	}

	if !slices.Contains(supportedArray, b.BuildTool) {
		return fmt.Errorf("unsupported build tool: %s, it should be one of %s", b.BuildTool, supportedString)
	}

	return nil
}

func (b BuildConfig) Clone(url, ref string) error {
	// Clone repo only when source dir not exists.
	if !fileio.PathExists(b.PortConfig.SourceDir) {
		if strings.HasSuffix(url, ".git") {
			// Clone repo.
			command := fmt.Sprintf("git clone --branch %s %s %s --recursive", ref, url, b.PortConfig.SourceDir)
			title := fmt.Sprintf("[clone %s@%s]", b.PortConfig.LibName, b.PortConfig.LibVersion)
			if err := cmd.NewExecutor(title, command).Execute(); err != nil {
				return err
			}
		} else {
			// Check and repair resource.
			archiveName := filepath.Base(url)
			repair := fileio.NewDownloadRepair(url, archiveName, ".", b.PortConfig.TmpDir, b.PortConfig.DownloadedDir)
			if err := repair.CheckAndRepair(); err != nil {
				return err
			}

			// Move extracted files to source dir.
			entities, err := os.ReadDir(b.PortConfig.TmpDir)
			if err != nil || len(entities) == 0 {
				return fmt.Errorf("cannot find extracted files under tmp dir: %w", err)
			}
			if len(entities) == 1 {
				sourceDir := filepath.Join(b.PortConfig.TmpDir, entities[0].Name())
				if err := fileio.RenameDir(sourceDir, b.PortConfig.SourceDir); err != nil {
					return err
				}
			} else if len(entities) > 1 {
				if err := fileio.RenameDir(b.PortConfig.TmpDir, b.PortConfig.SourceDir); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (b BuildConfig) Patch() error {
	if len(b.Patches) == 0 {
		return nil
	}

	// Apply all patches.
	for _, patch := range b.Patches {
		patch = strings.TrimSpace(patch)
		if patch == "" {
			continue
		}

		// Check if patch file exists.
		patchPath := filepath.Join(b.PortConfig.PortsDir, b.PortConfig.LibName, patch)
		if !fileio.PathExists(patchPath) {
			return fmt.Errorf("patch file %s doesn't exists", patchPath)
		}

		// Apply patch (linux patch or git patch).
		if err := cmd.ApplyPatch(b.PortConfig.SourceDir, patchPath); err != nil {
			return err
		}
	}

	return nil
}

func (b *BuildConfig) Install(url, ref, buildType string) error {
	// Check if system tool is already installed.
	if err := b.checkSystemTools(); err != nil {
		return err
	}

	// Clean repo if possible.
	if err := cmd.CleanRepo(b.PortConfig.SourceDir); err != nil {
		return fmt.Errorf("clean repo failed: %s", err)
	}

	// Set cross tool in environment for cross compiling.
	if b.AsDev {
		b.PortConfig.CrossTools.ClearEnvs()
	} else {
		b.PortConfig.CrossTools.SetEnvs()
	}

	// Replace placeholders with real value, like ${HOST}, ${SYSROOT} etc.
	b.buildSystem.fillPlaceHolders()

	// Some third-party need extra environment variables.
	if err := b.buildSystem.appendBuildEnvs(); err != nil {
		return err
	}
	defer b.buildSystem.removeBuildEnvs()

	if err := b.buildSystem.Clone(url, ref); err != nil {
		return err
	}
	if err := b.buildSystem.Patch(); err != nil {
		return err
	}
	if err := b.buildSystem.fixConfigure(); err != nil {
		return err
	}
	if err := b.buildSystem.Configure(buildType); err != nil {
		return err
	}

	if err := b.buildSystem.Build(); err != nil {
		// Some third-party need extra steps to fix build.
		// For example: nspr.
		if len(b.FixBuild.Scripts) > 0 {
			if err := b.buildSystem.fixBuild(); err != nil {
				return err
			}

			if err := b.buildSystem.Build(); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	if err := b.buildSystem.Install(); err != nil {
		return err
	}

	// Change pc file's prefix as the installed directory.
	var prefix string
	if !b.AsDev {
		prefix = strings.TrimPrefix(b.PortConfig.InstalledDir, b.PortConfig.WorkspaceDir)
	}
	if err := fixupPkgConfig(b.PortConfig.PackageDir, prefix); err != nil {
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
		cmakeConfig.SystemName = b.PortConfig.CrossTools.SystemName
		cmakeConfig.Libname = b.PortConfig.LibName
		cmakeConfig.BuildType = buildType
		if err := cmakeConfig.Generate(b.PortConfig.PackageDir); err != nil {
			return err
		}
	}

	// Create a symblink in the sysroot that points to the installed directory,
	// then the pc file would be found by other libraries.
	src := filepath.Dir(b.PortConfig.InstalledDir)
	dest := filepath.Join(b.PortConfig.CrossTools.RootFS, "installed")

	return b.checkInstalledSymblink(src, dest)
}

func (b *BuildConfig) InitBuildSystem() error {
	switch b.BuildTool {
	case "gyp":
		b.buildSystem = NewGyp(*b)
	case "cmake":
		b.buildSystem = NewCMake(*b, "")
	case "ninja":
		b.buildSystem = NewNinja(*b)
	case "makefiles":
		b.buildSystem = NewMakefiles(*b)
	case "meson":
		b.buildSystem = NewMeson(*b)
	case "b2":
		b.buildSystem = NewB2(*b)
	case "bazel":
		b.buildSystem = NewBazel(*b)
	default:
		return fmt.Errorf("unsupported build system: %s", b.BuildTool)
	}

	return nil
}

func (b BuildConfig) BuildSystem() BuildSystem {
	return b.buildSystem
}

// checkInstalledSymblink We create a symblink in the sysroot that points to the installed directory,
// then the pc file would be found by other libraries.
func (b BuildConfig) checkInstalledSymblink(src, dest string) error {
	// Convenient function to create a relative symblink.
	createSymblink := func(src, dest string) error {
		relPath, err := filepath.Rel(filepath.Dir(dest), src)
		if err != nil {
			return fmt.Errorf("failed to compute relative path: %w", err)
		}
		if err := os.Symlink(relPath, dest); err != nil {
			return fmt.Errorf("failed to create symlink: %v", err)
		}
		return nil
	}

	// Check if the symblink exists.
	info, err := os.Lstat(dest)
	if err != nil {
		if os.IsNotExist(err) {
			return createSymblink(src, dest)
		}
		return fmt.Errorf("failed to checking symlink: %v", err)
	}

	// Check the symblink target.
	if info.Mode()&os.ModeSymlink != 0 {
		// Read the target of the symlink.
		realTarget, err := os.Readlink(dest)
		if err != nil {
			return fmt.Errorf("failed to read symlink target: %v", err)
		}

		// If symlink is broken or points to the wrong target, remove it and recreate.
		if realTarget != src {
			if err := os.Remove(dest); err != nil {
				return fmt.Errorf("failed to remove broken symlink: %v", err)
			}
			return createSymblink(src, dest)
		}

		return nil
	}

	// Removeit if it's not a symlink.
	if err = os.Remove(dest); err != nil {
		return fmt.Errorf("failed to remove non-symlink: %v", err)
	}
	return createSymblink(src, dest)
}

func (b BuildConfig) checkSystemTools() error {
	var missing []string
	for _, tool := range b.SystemTools {
		tool = strings.TrimSpace(tool)
		if tool == "" {
			continue
		}

		installed, err := cmd.IsLibraryInstalled(tool)
		if err != nil {
			return err
		}

		if !installed {
			missing = append(missing, tool)
		}
	}

	if len(missing) > 0 {
		var summary string
		if len(missing) == 1 {
			summary = fmt.Sprintf("The system tool for `%s` is not installed", missing[0])
		} else if len(missing) == 2 {
			summary = fmt.Sprintf("The system tool for `%s` and `%s` are not installed", missing[0], missing[1])
		} else {
			summary = "The system tool for `" + strings.Join(missing[:len(missing)-1], "`, `") + " and `" + missing[len(missing)-1] + "` are not installed"
		}

		joined := strings.Join(missing, " ")
		if runtime.GOOS == "linux" {
			return fmt.Errorf("%s,\n    Please install it with `sudo apt install %s`", summary, joined)
		} else if runtime.GOOS == "windows" {
			return fmt.Errorf("%s,\n    Please install it with `pacman -S %s` in MSYS2", joined, joined)
		} else if runtime.GOOS == "darwin" {
			return fmt.Errorf("%s,\n    Please install it with `brew install %s`", joined, joined)
		}
	}

	return nil
}

func (b BuildConfig) fixConfigure() error {
	for _, script := range b.FixConfigure.Scripts {
		script = strings.TrimSpace(script)
		if script == "" {
			continue
		}

		// Replace placeholders with real value.
		script = b.replaceHolders(script)
		workDir := b.replaceHolders(b.FixConfigure.WorkDir)

		title := fmt.Sprintf("[before confiure %s]", b.PortConfig.LibName)
		executor := cmd.NewExecutor(title, script)
		executor.SetWorkDir(workDir)
		if err := executor.Execute(); err != nil {
			return err
		}
	}

	return nil
}

func (b BuildConfig) fixBuild() error {
	for _, script := range b.FixBuild.Scripts {
		script = strings.TrimSpace(script)
		if script == "" {
			continue
		}

		// Replace placeholders with real value.
		script = b.replaceHolders(script)
		workDir := b.replaceHolders(b.FixBuild.WorkDir)

		title := fmt.Sprintf("[fix build %s]", b.PortConfig.LibName)
		executor := cmd.NewExecutor(title, script)
		executor.SetWorkDir(workDir)
		if err := executor.Execute(); err != nil {
			return err
		}
	}

	return nil
}

func (b *BuildConfig) appendBuildEnvs() error {
	b.environment.Backup()

	for _, item := range b.EnvVars {
		item = strings.TrimSpace(item)

		index := strings.Index(item, "=")
		if index == -1 {
			return fmt.Errorf("invalid env var: %s", item)
		}

		key := strings.TrimSpace(item[:index])
		value := strings.TrimSpace(item[index+1:])
		value = b.replaceHolders(value)

		switch key {
		case "CPATH":
			current := os.Getenv(key)
			if strings.TrimSpace(current) == "" {
				os.Setenv(key, value)
			} else {
				os.Setenv(key, value+string(os.PathListSeparator)+current)
			}

		case "CFLAGS", "CXXFLAGS":
			// buildenv can wrap CFLAGS and CXXFLAGS, so we need to remove them.
			value = strings.ReplaceAll(value, "${CFLAGS}", "")
			value = strings.ReplaceAll(value, "${CXXFLAGS}", "")

			current := os.Getenv(key)
			if strings.TrimSpace(current) == "" {
				os.Setenv(key, strings.TrimSpace(value))
			} else {
				os.Setenv(key, fmt.Sprintf("%s %s", current, value))
			}

		default:
			os.Setenv(key, value)
		}
	}

	// Make sure installed libaries can be found via pkg-config during compiling.
	if b.AsDev {
		var pkgConfigs = []string{
			fmt.Sprintf("%s/lib/pkgconfig", b.PortConfig.InstalledDir),
			fmt.Sprintf("%s/share/pkgconfig", b.PortConfig.InstalledDir),
		}
		os.Setenv("PKG_CONFIG_PATH", strings.Join(pkgConfigs, string(os.PathListSeparator)))
		os.Setenv("PKG_CONFIG_SYSROOT_DIR", b.PortConfig.InstalledDir)
	} else {
		if b.PortConfig.CrossTools.RootFS != "" {
			os.Setenv("SYSROOT", b.PortConfig.CrossTools.RootFS)

			// Add extra header dirs into search path.
			var extraHeaderDirsString = func() string {
				var result []string
				for _, path := range b.PortConfig.ExtraHeaderDirs {
					result = append(result, "-I"+filepath.Join(b.PortConfig.CrossTools.RootFS, path))
				}
				return strings.Join(result, " ")
			}
			joinedDirs := extraHeaderDirsString()
			if joinedDirs == "" {
				env.AppendEnv("CFLAGS", fmt.Sprintf("--sysroot=%s", b.PortConfig.CrossTools.RootFS))
				env.AppendEnv("CXXFLAGS", fmt.Sprintf("--sysroot=%s", b.PortConfig.CrossTools.RootFS))
			} else {
				env.AppendEnv("CFLAGS", fmt.Sprintf("--sysroot=%s %s", b.PortConfig.CrossTools.RootFS, joinedDirs))
				env.AppendEnv("CXXFLAGS", fmt.Sprintf("--sysroot=%s %s", b.PortConfig.CrossTools.RootFS, joinedDirs))
			}

			// Add extra lib dirs into search path.
			var extraLibDirsString = func() string {
				var result []string
				for _, path := range b.PortConfig.ExtraLibDirs {
					result = append(result, filepath.Join(b.PortConfig.CrossTools.RootFS, path))
				}
				return strings.Join(result, string(os.PathListSeparator))
			}
			env.AppendEnv("LDFLAGS", fmt.Sprintf("--sysroot=%s", b.PortConfig.CrossTools.RootFS))
			env.AppendRPathLink(extraLibDirsString())

			var pkgConfigs = []string{
				fmt.Sprintf("%s/installed/%s/lib/pkgconfig", b.PortConfig.CrossTools.RootFS, b.PortConfig.InstalledFolder),
				fmt.Sprintf("%s/installed/%s/share/pkgconfig", b.PortConfig.CrossTools.RootFS, b.PortConfig.InstalledFolder),
				os.Getenv("PKG_CONFIG_PATH"),
			}
			os.Setenv("PKG_CONFIG_PATH", strings.Join(pkgConfigs, string(os.PathListSeparator)))
			os.Setenv("PKG_CONFIG_SYSROOT_DIR", b.PortConfig.CrossTools.RootFS)
		} else {
			var pkgConfigs = []string{
				fmt.Sprintf("%s/lib/pkgconfig", b.PortConfig.InstalledDir),
				fmt.Sprintf("%s/share/pkgconfig", b.PortConfig.InstalledDir),
			}
			os.Setenv("PKG_CONFIG_PATH", strings.Join(pkgConfigs, string(os.PathListSeparator)))
			os.Setenv("PKG_CONFIG_SYSROOT_DIR", b.PortConfig.InstalledDir)
		}

		// Append "--sysroot=" for cross compile.
		installedHeaderDir := fmt.Sprintf("%s/installed/%s/include", b.PortConfig.CrossTools.RootFS, b.PortConfig.InstalledFolder)
		env.AppendEnv("CFLAGS", fmt.Sprintf("-I%s", installedHeaderDir))
		env.AppendEnv("CXXFLAGS", fmt.Sprintf("-I%s", installedHeaderDir))

		// Append rpath-link.
		env.AppendRPathLink(filepath.Join(b.PortConfig.InstalledDir, "lib"))
	}

	return nil
}

func (b BuildConfig) removeBuildEnvs() error {
	b.environment.Rollback()
	return nil
}

// fillPlaceHolders Replace placeholders with real paths and values.
func (b *BuildConfig) fillPlaceHolders() {
	for index, argument := range b.Options {
		if strings.Contains(argument, "${HOST}") {
			if b.AsDev {
				b.Options = slices.Delete(b.Options, index, 1)
			} else {
				b.Options[index] = strings.ReplaceAll(argument, "${HOST}", b.PortConfig.CrossTools.Host)
			}
		}

		if strings.Contains(argument, "${SYSTEM_NAME}") {
			if b.AsDev {
				b.Options = slices.Delete(b.Options, index, 1)
			} else {
				b.Options[index] = strings.ReplaceAll(argument, "${SYSTEM_NAME}", strings.ToLower(b.PortConfig.CrossTools.SystemName))
			}
		}

		if strings.Contains(argument, "${SYSTEM_PROCESSOR}") {
			if b.AsDev {
				b.Options = slices.Delete(b.Options, index, 1)
			} else {
				b.Options[index] = strings.ReplaceAll(argument, "${SYSTEM_PROCESSOR}", b.PortConfig.CrossTools.SystemProcessor)
			}
		}

		if strings.Contains(argument, "${SYSROOT}") {
			if b.AsDev {
				b.Options = slices.Delete(b.Options, index, 1)
			} else {
				b.Options[index] = strings.ReplaceAll(argument, "${SYSROOT}", b.PortConfig.CrossTools.RootFS)
			}
		}

		if strings.Contains(argument, "${CROSS_PREFIX}") {
			if b.AsDev {
				b.Options = slices.Delete(b.Options, index, 1)
			} else {
				b.Options[index] = strings.ReplaceAll(argument, "${CROSS_PREFIX}", b.PortConfig.CrossTools.ToolchainPrefix)
			}
		}

		if strings.Contains(argument, "${INSTALLED_DIR}") {
			b.Options[index] = strings.ReplaceAll(argument, "${INSTALLED_DIR}", b.PortConfig.InstalledDir)
		}

		if strings.Contains(argument, "${SOURCE_DIR}") {
			b.Options[index] = strings.ReplaceAll(argument, "${SOURCE_DIR}", b.PortConfig.SourceDir)
		}
	}
}

func (b BuildConfig) setBuildType(buildType string) {
	// Remove all -g and -O flags.
	cflags := strings.Split(os.Getenv("CFLAGS"), " ")
	cflags = slices.DeleteFunc(cflags, func(element string) bool {
		element = strings.TrimSpace(element)
		return element == "-g" || element == "-O"
	})

	cxxflags := strings.Split(os.Getenv("CXXFLAGS"), " ")
	cxxflags = slices.DeleteFunc(cxxflags, func(element string) bool {
		element = strings.TrimSpace(element)
		return element == "-g" || element == "-O"
	})

	if b.AsDev {
		// Set -O3 for dev.
		cflags = append(cflags, "-O3")
		cxxflags = append(cxxflags, "-O3")
		os.Setenv("CFLAGS", strings.Join(cflags, " "))
		os.Setenv("CXXFLAGS", strings.Join(cxxflags, " "))
	} else {
		// Set -g for debug and -O3 for release.
		var flags string
		if strings.ToLower(buildType) == "debug" {
			flags = "-g"
		} else {
			flags = "-O3"
		}

		cflags = append(cflags, flags)
		cxxflags = append(cxxflags, flags)
		os.Setenv("CFLAGS", strings.Join(cflags, " "))
		os.Setenv("CXXFLAGS", strings.Join(cxxflags, " "))
	}
}

func (b BuildConfig) replaceHolders(content string) string {
	content = strings.ReplaceAll(content, "${HOST}", b.PortConfig.CrossTools.Host)
	content = strings.ReplaceAll(content, "${SYSTEM_NAME}", b.PortConfig.CrossTools.SystemName)
	content = strings.ReplaceAll(content, "${SYSTEM_PROCESSOR}", b.PortConfig.CrossTools.SystemProcessor)
	content = strings.ReplaceAll(content, "${SYSROOT}", b.PortConfig.CrossTools.RootFS)
	content = strings.ReplaceAll(content, "${CROSS_PREFIX}", b.PortConfig.CrossTools.ToolchainPrefix)
	content = strings.ReplaceAll(content, "${INSTALLED_DIR}", b.PortConfig.InstalledDir)
	content = strings.ReplaceAll(content, "${SOURCE_DIR}", b.PortConfig.SourceDir)
	content = strings.ReplaceAll(content, "${BUILD_DIR}", b.PortConfig.BuildDir)
	content = strings.ReplaceAll(content, "${PACKAGE_DIR}", b.PortConfig.PackageDir)
	content = strings.ReplaceAll(content, "${DEV_DIR}", filepath.Join(filepath.Dir(b.PortConfig.InstalledDir), "dev"))
	return content
}

func (b BuildConfig) getLogPath(suffix string) string {
	parentDir := filepath.Dir(b.PortConfig.BuildDir)
	fileName := filepath.Base(b.PortConfig.BuildDir) + fmt.Sprintf("-%s.log", suffix)
	return filepath.Join(parentDir, fileName)
}
