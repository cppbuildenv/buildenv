package buildsystem

import (
	"buildenv/generator"
	"buildenv/pkg/cmd"
	"buildenv/pkg/fileio"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var supportedArray = []string{"b2", "bazel", "cmake", "gyp", "makefiles", "meson", "ninja"}

const supportedString = "b2, bazel, cmake, gyp, makefiles, meson, ninja"

type PortConfig struct {
	LibName    string // like: `ffmpeg`
	LibVersion string // like: `4.4`

	// Internal fields
	CrossTools    CrossTools // cross tools like CC, CXX, FC, RANLIB, AR, LD, NM, OBJDUMP, STRIP
	PortsDir      string     // ${buildenv}/ports
	DownloadedDir string     // ${buildenv}/downloads
	SourceDir     string     // for example: ${buildenv}/buildtrees/ffmpeg/src
	SourceFolder  string     // Some thirdpartys' source code is not in the root folder, so we need to specify it.
	BuildDir      string     // for example: ${buildenv}/buildtrees/ffmpeg/x86_64-linux-20.04-Release
	PackageDir    string     // for example: ${buildenv}/packages/ffmpeg-3.4.13-x86_64-linux-20.04-Release
	InstalledDir  string     // for example: ${buildenv}/installed/x86_64-linux-20.04-Release
	WithSubmodule bool       // if true, clone submodule when clone repository
	JobNum        int        // number of jobs to run in parallel
	TmpDir        string     // for example: ${buildenv}/downloaded/tmp
}

type BuildSystem interface {
	Clone(repoUrl, repoRef string) error
	Patch() error
	Configure(buildType string) error
	Build() error
	Install() error
	PackageFiles(packageDir, platformName, projectName, buildType string) ([]string, error)

	fixConfigure() error
	fixBuild() error // Some thirdpartys need extra steps to fix build, for example: nspr.
	appendBuildEnvs() error
	removeBuildEnvs() error
	fillPlaceHolders()
	setBuildType(buildType string)
	ensureDependencyPaths()
	getLogPath(suffix string) string
}

// CrossTools same with `Toolchain` in config/toolchain.go
// redefine to avoid import cycle.
type CrossTools struct {
	SystemName      string
	SystemProcessor string
	Host            string
	RootFS          string
	ToolchainPrefix string
	CC              string
	CXX             string
	FC              string
	RANLIB          string
	AR              string
	LD              string
	NM              string
	OBJDUMP         string
	STRIP           string
	Native          bool
}

func (c CrossTools) SetEnvs() {
	if c.Native {
		return
	}

	// Set env vars only for cross compiling.
	rootfs := os.Getenv("SYSROOT")
	os.Setenv("TOOLCHAIN_PREFIX", c.ToolchainPrefix)
	os.Setenv("HOST", c.Host)
	os.Setenv("CC", fmt.Sprintf("%s --sysroot=%s", c.CC, rootfs))
	os.Setenv("CXX", fmt.Sprintf("%s --sysroot=%s", c.CXX, rootfs))

	if c.FC != "" {
		os.Setenv("FC", c.FC)
	}

	if c.RANLIB != "" {
		os.Setenv("RANLIB", c.RANLIB)
	}

	if c.AR != "" {
		os.Setenv("AR", c.AR)
	}

	if c.LD != "" {
		os.Setenv("LD", fmt.Sprintf("%s --sysroot=%s", c.LD, rootfs))
	}

	if c.NM != "" {
		os.Setenv("NM", c.NM)
	}

	if c.OBJDUMP != "" {
		os.Setenv("OBJDUMP", c.OBJDUMP)
	}

	if c.STRIP != "" {
		os.Setenv("STRIP", c.STRIP)
	}
}

func (CrossTools) ClearEnvs() {
	os.Unsetenv("TOOLCHAIN_PREFIX")
	os.Unsetenv("HOST")
	os.Unsetenv("CC")
	os.Unsetenv("CXX")
	os.Unsetenv("FC")
	os.Unsetenv("RANLIB")
	os.Unsetenv("AR")
	os.Unsetenv("LD")
	os.Unsetenv("NM")
	os.Unsetenv("OBJDUMP")
	os.Unsetenv("STRIP")
}

type scriptWork struct {
	Scripts []string `json:"scripts"`
	WorkDir string   `json:"work_dir"`
}

type BuildConfig struct {
	Pattern        string     `json:"pattern"`
	BuildTool      string     `json:"build_tool"`
	LibraryType    string     `json:"library_type"`
	EnvVars        []string   `json:"env_vars"`
	FixConfigure   scriptWork `json:"fix_configure"`
	FixBuild       scriptWork `json:"fix_build"`
	Patches        []string   `json:"patches"`
	Arguments      []string   `json:"arguments"`
	Depedencies    []string   `json:"dependencies"`
	DevDepedencies []string   `json:"dev_dependencies"`
	CMakeConfig    string     `json:"cmake_config"`

	// Internal fields
	buildSystem BuildSystem `json:"-"`
	PortConfig  PortConfig  `json:"-"`
	AsDev       bool        `json:"-"`
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

func (b BuildConfig) Clone(url, version string) error {
	// Clone repo only when source dir not exists.
	if !fileio.PathExists(b.PortConfig.SourceDir) {
		if strings.HasSuffix(url, ".git") {
			// Clone repo.
			var command string
			if b.PortConfig.WithSubmodule {
				command = fmt.Sprintf("git clone --branch --recursive %s %s %s", version, url, b.PortConfig.SourceDir)
			} else {
				command = fmt.Sprintf("git clone --branch %s %s %s", version, url, b.PortConfig.SourceDir)
			}
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

	// Change to source dir.
	if err := os.Chdir(b.PortConfig.SourceDir); err != nil {
		return err
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
			return fmt.Errorf("patch file %s not exists", patchPath)
		}

		// Apply patch (linux patch or git patch).
		if err := cmd.ApplyPatch(b.PortConfig.SourceDir, patchPath); err != nil {
			return err
		}
	}

	return nil
}

func (b *BuildConfig) Install(url, version, buildType string) error {
	// Replace placeholders with real value, like ${HOST}, ${SYSROOT} etc.
	b.buildSystem.fillPlaceHolders()

	// Some third-party need extra environment variables.
	if err := b.buildSystem.appendBuildEnvs(); err != nil {
		return err
	}
	defer b.buildSystem.removeBuildEnvs()

	// Make sure depedencies libs can be found by current lib.
	b.buildSystem.ensureDependencyPaths()

	if err := b.buildSystem.Clone(url, version); err != nil {
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
		cmakeConfig.SystemName = b.PortConfig.CrossTools.SystemName
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

		if b.AsDev {
			files = append(files, filepath.Join("dev", relativePath))
		} else {
			platformProject := fmt.Sprintf("%s^%s^%s", platformName, projectName, buildType)
			files = append(files, filepath.Join(platformProject, relativePath))
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return files, nil
}

func (b BuildConfig) BuildSystem() BuildSystem {
	return b.buildSystem
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

func (b BuildConfig) appendBuildEnvs() error {
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
		case "PKG_CONFIG_PATH", "PATH":
			value = fmt.Sprintf("%s%s%s", value, string(os.PathListSeparator), os.Getenv(key))
			os.Setenv(key, value)

		case "CFLAGS", "CXXFLAGS":
			current := os.Getenv(key)
			if strings.TrimSpace(current) == "" {
				os.Setenv(key, value)
			} else {
				os.Setenv(key, fmt.Sprintf("%s %s", current, value))
			}
		}
	}

	return nil
}

func (b BuildConfig) removeBuildEnvs() error {
	for _, item := range b.EnvVars {
		item = strings.TrimSpace(item)
		index := strings.Index(item, "=")
		if index == -1 {
			return fmt.Errorf("invalid env var: %s", item)
		}

		key := strings.TrimSpace(item[:index])
		value := strings.TrimSpace(item[index+1:])

		switch key {
		case "CFLAGS", "CXXFLAGS":
			flagsValue := strings.ReplaceAll(os.Getenv(key), value, "")
			if strings.TrimSpace(flagsValue) == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, flagsValue)
			}

		case "PKG_CONFIG_PATH", "PATH":
			parts := strings.Split(os.Getenv("PKG_CONFIG_PATH"), string(os.PathListSeparator))
			// Remove the value from the slice.
			for i, part := range parts {
				if part == value {
					parts = append(parts[:i], parts[i+1:]...)
					break
				}
			}

			// Reconstruct the PKG_CONFIG_PATH string.
			if len(parts) == 0 {
				os.Unsetenv("PKG_CONFIG_PATH")
			} else {
				os.Setenv("PKG_CONFIG_PATH", strings.Join(parts, string(os.PathListSeparator)))
			}
		}
	}

	return nil
}

// fillPlaceHolders Replace placeholders with real paths and values.
func (b *BuildConfig) fillPlaceHolders() {
	for index, argument := range b.Arguments {
		if strings.Contains(argument, "${HOST}") {
			if b.AsDev {
				b.Arguments = slices.Delete(b.Arguments, index, 1)
			} else {
				b.Arguments[index] = strings.ReplaceAll(argument, "${HOST}", b.PortConfig.CrossTools.Host)
			}
		}

		if strings.Contains(argument, "${SYSTEM_NAME}") {
			if b.AsDev {
				b.Arguments = slices.Delete(b.Arguments, index, 1)
			} else {
				b.Arguments[index] = strings.ReplaceAll(argument, "${SYSTEM_NAME}", strings.ToLower(b.PortConfig.CrossTools.SystemName))
			}
		}

		if strings.Contains(argument, "${SYSTEM_PROCESSOR}") {
			if b.AsDev {
				b.Arguments = slices.Delete(b.Arguments, index, 1)
			} else {
				b.Arguments[index] = strings.ReplaceAll(argument, "${SYSTEM_PROCESSOR}", b.PortConfig.CrossTools.SystemProcessor)
			}
		}

		if strings.Contains(argument, "${SYSROOT}") {
			if b.AsDev {
				b.Arguments = slices.Delete(b.Arguments, index, 1)
			} else {
				b.Arguments[index] = strings.ReplaceAll(argument, "${SYSROOT}", b.PortConfig.CrossTools.RootFS)
			}
		}

		if strings.Contains(argument, "${CROSS_PREFIX}") {
			if b.AsDev {
				b.Arguments = slices.Delete(b.Arguments, index, 1)
			} else {
				b.Arguments[index] = strings.ReplaceAll(argument, "${CROSS_PREFIX}", b.PortConfig.CrossTools.ToolchainPrefix)
			}
		}

		if strings.Contains(argument, "${INSTALLED_DIR}") {
			if b.AsDev {
				b.Arguments = slices.Delete(b.Arguments, index, 1)
			} else {
				b.Arguments[index] = strings.ReplaceAll(argument, "${INSTALLED_DIR}", b.PortConfig.InstalledDir)
			}
		}

		if strings.Contains(argument, "${SOURCE_DIR}") {
			if b.AsDev {
				b.Arguments = slices.Delete(b.Arguments, index, 1)
			} else {
				b.Arguments[index] = strings.ReplaceAll(argument, "${SOURCE_DIR}", b.PortConfig.SourceDir)
			}
		}
	}
}

func (b BuildConfig) setBuildType(buildType string) {
	// Remove all -g and -O flags.
	cflags := strings.Split(os.Getenv("CFLAGS"), " ")
	cflags = slices.DeleteFunc(cflags, func(element string) bool {
		return strings.Contains(element, "-g") || strings.Contains(element, "-O")
	})

	cxxflags := strings.Split(os.Getenv("CXXFLAGS"), " ")
	cxxflags = slices.DeleteFunc(cxxflags, func(element string) bool {
		return strings.Contains(element, "-g") || strings.Contains(element, "-O")
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

// ensureDependencyPaths Sometimes libs not in sysroot cannot be found,
// we need to set CFLAGS, CXXFLAGS, LDFLAGS to make sure these third-party
// libaries that installed in installed dir can be found.
func (b BuildConfig) ensureDependencyPaths() {
	installedDir := b.PortConfig.InstalledDir
	cflags := os.Getenv("CFLAGS")
	cxxflags := os.Getenv("CXXFLAGS")
	ldflags := os.Getenv("LDFLAGS")

	if strings.TrimSpace(cflags) == "" {
		os.Setenv("CFLAGS", fmt.Sprintf("-I%s/include", installedDir))
	} else {
		os.Setenv("CFLAGS", fmt.Sprintf("-I%s/include", installedDir)+" "+cflags)
	}
	if strings.TrimSpace(cxxflags) == "" {
		os.Setenv("CXXFLAGS", fmt.Sprintf("-I%s/include", installedDir))
	} else {
		os.Setenv("CXXFLAGS", fmt.Sprintf("-I%s/include", installedDir)+" "+cxxflags)
	}
	if strings.TrimSpace(ldflags) == "" {
		os.Setenv("LDFLAGS", fmt.Sprintf("-Wl,-rpath-link,%s/lib", installedDir))
	} else {
		os.Setenv("LDFLAGS", fmt.Sprintf("-Wl,-rpath-link,%s/lib", installedDir)+" "+ldflags)
	}

	// Append $PKG_CONFIG_PATH with pkgconfig path that in installed dir.
	pkgConfigPath := os.Getenv("PKG_CONFIG_PATH")
	if strings.TrimSpace(pkgConfigPath) == "" {
		os.Setenv("PKG_CONFIG_PATH", installedDir+"/lib/pkgconfig")
	} else {
		os.Setenv("PKG_CONFIG_PATH", installedDir+"/lib/pkgconfig"+string(os.PathListSeparator)+pkgConfigPath)
	}

	// We assume that pkg-config's sysroot is installedDir and change all pc file's prefix as "/".
	os.Setenv("PKG_CONFIG_SYSROOT_DIR", installedDir)
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
	return content
}

func (b BuildConfig) getLogPath(suffix string) string {
	parentDir := filepath.Dir(b.PortConfig.BuildDir)
	fileName := filepath.Base(b.PortConfig.BuildDir) + fmt.Sprintf("-%s.log", suffix)
	return filepath.Join(parentDir, fileName)
}
