package config

import (
	"buildenv/buildsystem"
	"buildenv/pkg/color"
	"buildenv/pkg/fileio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Port struct {
	Url           string                    `json:"url"`
	Name          string                    `json:"name"`
	Version       string                    `json:"version"`
	WithSubmodule bool                      `json:"with_submodule"`
	SourceFolder  string                    `json:"source_folder,omitempty"`
	BuildConfigs  []buildsystem.BuildConfig `json:"build_configs"`

	// Internal fields.
	ctx       Context `json:"-"`
	stateFile string  `json:"-"` // Used to record installed state
	AsSubDep  bool    `json:"-"`
	AsDev     bool    `json:"-"`
}

func (p Port) NameVersion() string {
	return p.Name + "@" + p.Version
}

func (p *Port) Init(ctx Context, portPath string) error {
	p.ctx = ctx

	// Add file suffix and prefix if not exists.
	if !strings.HasSuffix(portPath, ".json") {
		portPath += ".json"
	}
	if !strings.HasPrefix(portPath, Dirs.PortsDir) {
		portPath = filepath.Join(Dirs.PortsDir, portPath)
	}

	// Read name and version.
	portPath = strings.ReplaceAll(portPath, "@", "/")
	if !fileio.PathExists(portPath) {
		version := fileio.FileBaseName(portPath)
		name := fileio.FileBaseName(filepath.Dir(portPath))
		nameVersion := name + "@" + version

		if p.AsSubDep {
			return fmt.Errorf("sub depedency port %s does not exists", nameVersion)
		} else {
			return fmt.Errorf("port %s does not exists", nameVersion)
		}
	}

	// Decode JSON.
	bytes, err := os.ReadFile(portPath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bytes, p); err != nil {
		return err
	}

	// Info file: used to record installed state.
	nameVersion := p.NameVersion()

	var (
		installedFolder string
		packageFolder   string
		buildFolder     string
	)
	if p.AsDev {
		packageFolder = nameVersion
		installedFolder = "dev"
		buildFolder = filepath.Join(nameVersion, "dev")
		p.stateFile = filepath.Join(Dirs.InstalledDir, "buildenv", "info", nameVersion+"^dev.list")
	} else {
		platformProject := fmt.Sprintf("%s^%s^%s", ctx.Platform().Name, ctx.Project().Name, ctx.BuildType())
		packageFolder = fmt.Sprintf("%s^%s^%s^%s", nameVersion, ctx.Platform().Name, ctx.Project().Name, ctx.BuildType())
		installedFolder = fmt.Sprintf("%s^%s^%s", ctx.Platform().Name, ctx.Project().Name, ctx.BuildType())
		buildFolder = filepath.Join(nameVersion, fmt.Sprintf("%s^%s^%s", ctx.Platform().Name, ctx.Project().Name, ctx.BuildType()))
		p.stateFile = filepath.Join(Dirs.InstalledDir, "buildenv", "info", nameVersion+"^"+platformProject+".list")
	}

	portConfig := buildsystem.PortConfig{
		CrossTools:    p.buildCrossTools(),
		JobNum:        ctx.JobNum(),
		LibName:       p.Name,
		LibVersion:    p.Version,
		SourceFolder:  p.SourceFolder,
		PortsDir:      Dirs.PortsDir,
		DownloadedDir: Dirs.DownloadedDir,
		SourceDir:     filepath.Join(Dirs.WorkspaceDir, "buildtrees", nameVersion, "src"),
		BuildDir:      filepath.Join(Dirs.WorkspaceDir, "buildtrees", buildFolder),
		PackageDir:    filepath.Join(Dirs.WorkspaceDir, "packages", packageFolder),
		InstalledDir:  filepath.Join(Dirs.InstalledDir, installedFolder),
		WithSubmodule: p.WithSubmodule,
		TmpDir:        filepath.Join(Dirs.DownloadedDir, "tmp"),
	}

	if len(p.BuildConfigs) > 0 {
		for index := range p.BuildConfigs {
			p.BuildConfigs[index].PortConfig = portConfig
			p.BuildConfigs[index].AsDev = p.AsDev

			// Merge project override ports.
			p.mergeBuildConfig(&p.BuildConfigs[index], ctx.Project().OverridePorts)
		}
	}

	return nil
}

func (p *Port) Validate() error {
	if p.Url == "" {
		return fmt.Errorf("url of %s is empty", p.Name)
	}

	if p.Name == "" {
		return fmt.Errorf("name of %s is empty", p.Name)
	}

	if p.Version == "" {
		return fmt.Errorf("version of %s is empty", p.Name)
	}

	for _, config := range p.BuildConfigs {
		if !p.MatchPattern(config.Pattern) {
			continue
		}

		if err := config.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (p Port) Installed() bool {
	if !fileio.PathExists(p.stateFile) {
		return false
	}

	// File can be read?
	bytes, err := os.ReadFile(p.stateFile)
	if err != nil {
		return false
	}

	// No content?
	if len(bytes) == 0 {
		return false
	}

	return true
}

func (p Port) Write(portPath string) error {
	p.BuildConfigs = []buildsystem.BuildConfig{}
	bytes, err := json.MarshalIndent(p, "", "    ")
	if err != nil {
		return err
	}

	// Check if tool exists.
	if fileio.PathExists(portPath) {
		return fmt.Errorf("%s is already exists", portPath)
	}

	// Make sure the parent directory exists.
	parentDir := filepath.Dir(portPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(portPath, bytes, os.ModePerm)
}

func (p Port) Install(silentMode bool) error {
	var installedDir string
	if p.AsDev {
		installedDir = filepath.Join(Dirs.WorkspaceDir, "installed", "dev")
	} else {
		installedDir = filepath.Join(Dirs.WorkspaceDir, "installed", p.ctx.Platform().Name+"-"+p.ctx.BuildType())
	}
	if p.Installed() {
		if !silentMode {
			title := color.Sprintf(color.Green, "\n[✔] ---- Port: %s\n", p.NameVersion())
			fmt.Printf("%sLocation: %s\n", title, installedDir)
		}
		return nil
	}

	// No config found, download and deploy it.
	if len(p.BuildConfigs) == 0 {
		downloadedDir := filepath.Join(Dirs.WorkspaceDir, "downloads")
		if err := p.downloadAndDeploy(p.Url, installedDir, downloadedDir); err != nil {
			return err
		}
	}

	// Find matched config and init build system.
	var matchedConfig *buildsystem.BuildConfig
	for _, config := range p.BuildConfigs {
		if p.MatchPattern(config.Pattern) {
			if err := config.InitBuildSystem(); err != nil {
				return err
			}
			matchedConfig = &config
			break
		}
	}
	if matchedConfig == nil {
		return fmt.Errorf("no matching build_config found to build")
	}

	var installedFrom string

	// Install from package dir.
	if fileio.PathExists(matchedConfig.PortConfig.PackageDir) {
		if err := p.installFromPackage(matchedConfig); err != nil {
			return err
		}
		installedFrom = "package"
	} else {
		// Try to install from cache.
		installed, fromDir, err := p.installFromCache(matchedConfig)
		if err != nil {
			return err
		}

		if installed {
			installedFrom = fmt.Sprintf("cache [%s]", fromDir)
		} else {
			// Remove build cache from buildtrees.
			platformProject := fmt.Sprintf("%s^%s^%s", p.ctx.Platform().Name, p.ctx.Project().Name, p.ctx.BuildType())
			logPathPrefix := filepath.Join(p.NameVersion(), platformProject)
			p.tryRemoveBuildCache(logPathPrefix)

			// Install from source when cache not found.
			if err := p.installFromSource(silentMode, matchedConfig); err != nil {
				return err
			}

			// Write package to cache dirs so that others can share installed libraries,
			// but only for none-dev lib.
			if !p.AsDev {
				for _, cacheDir := range p.ctx.CacheDirs() {
					if !cacheDir.Writable {
						continue
					}

					if err := cacheDir.Write(matchedConfig.PortConfig.PackageDir); err != nil {
						return err
					}
				}
			}

			installedFrom = "source"
		}

		// This will copy all install files into installed dir.
		if err := p.installFromPackage(matchedConfig); err != nil {
			return err
		}
	}

	// Write installed files info into its installation info list.
	if err := os.MkdirAll(filepath.Dir(p.stateFile), os.ModePerm); err != nil {
		return err
	}
	packageFiles, err := matchedConfig.BuildSystem().PackageFiles(
		matchedConfig.PortConfig.PackageDir,
		p.ctx.Platform().Name,
		p.ctx.Project().Name,
		p.ctx.BuildType(),
	)
	if err != nil {
		return err
	}
	if err := os.WriteFile(p.stateFile, []byte(strings.Join(packageFiles, "\n")), os.ModePerm); err != nil {
		return err
	}

	// Print install info when not in silent mode.
	if !silentMode {
		title := color.Sprintf(color.Green, "\n[✔] ---- Port: %s, installed from %s\n",
			p.NameVersion(), installedFrom)
		fmt.Printf("%sLocation: %s\n", title, installedDir)
	}

	return nil
}

func (p Port) MatchPattern(pattern string) bool {
	pattern = strings.TrimSpace(pattern)

	if pattern == "" || pattern == "*" {
		return true
	}

	platformName := p.ctx.Platform().Name
	if pattern[0] == '*' && pattern[len(pattern)-1] == '*' {
		return strings.Contains(platformName, pattern[1:len(pattern)-1])
	}

	if pattern[0] == '*' {
		return strings.HasSuffix(platformName, pattern[1:])
	}

	if pattern[len(pattern)-1] == '*' {
		return strings.HasPrefix(platformName, pattern[:len(pattern)-1])
	}

	return platformName == pattern
}

func (p *Port) mergeBuildConfig(portBuildConfig *buildsystem.BuildConfig, overrides map[string]buildsystem.BuildConfig) {
	if config, ok := overrides[p.NameVersion()]; ok {
		if config.LibraryType != "" {
			portBuildConfig.LibraryType = config.LibraryType
		}
		if len(config.EnvVars) > 0 {
			portBuildConfig.EnvVars = config.EnvVars
		}
		if config.Patches != nil {
			portBuildConfig.Patches = config.Patches
		}
		if len(config.Arguments) > 0 {
			portBuildConfig.Arguments = config.Arguments
		}
		portBuildConfig.Depedencies = config.Depedencies
		portBuildConfig.DevDepedencies = config.DevDepedencies
	}
}

func (p Port) installFromCache(matchedConfig *buildsystem.BuildConfig) (installed bool, cacheDir string, err error) {
	for _, cacheDir := range p.ctx.CacheDirs() {
		if !cacheDir.Readable {
			continue
		}

		ok, err := cacheDir.Read(
			p.ctx.Platform().Name,
			p.ctx.Project().Name,
			p.ctx.BuildType(),
			p.NameVersion()+".tar.gz",
			matchedConfig.PortConfig.PackageDir,
		)
		if err != nil {
			return false, "", err
		}
		if ok {
			return true, cacheDir.Dir, nil
		}
	}

	return false, "", nil
}

func (p Port) installFromSource(silentMode bool, buildConfig *buildsystem.BuildConfig) error {
	// 1. check and repair dev_dependencies.
	for _, item := range buildConfig.DevDepedencies {
		if strings.HasPrefix(item, p.Name) {
			return fmt.Errorf("%s's dev_dependencies contains circular dependency: %s", p.NameVersion(), item)
		}

		// Check and repair dependency.
		var port Port
		port.AsSubDep = true
		port.AsDev = true
		portPath := filepath.Join(Dirs.PortsDir, item+".json")
		if err := port.Init(p.ctx, portPath); err != nil {
			return err
		}
		if err := port.Install(silentMode); err != nil {
			return err
		}
	}

	// 2. check and repair dependencies.
	for _, item := range buildConfig.Depedencies {
		if strings.HasPrefix(item, p.Name) {
			return fmt.Errorf("%s's dependencies contains circular dependency: %s", p.NameVersion(), item)
		}

		// Check and repair dependency.
		var port Port
		port.AsSubDep = true
		portPath := filepath.Join(Dirs.PortsDir, item+".json")
		if err := port.Init(p.ctx, portPath); err != nil {
			return err
		}
		if err := port.Install(silentMode); err != nil {
			return err
		}
	}

	// Check and repair current port.
	if err := buildConfig.Install(p.Url, p.Version, p.ctx.BuildType()); err != nil {
		return err
	}

	return nil
}

func (p Port) installFromPackage(matchedConfig *buildsystem.BuildConfig) error {
	platformProject := fmt.Sprintf("%s^%s^%s", p.ctx.Platform().Name, p.ctx.Project().Name, p.ctx.BuildType())

	// First, we must check and repair dependency ports.
	for _, nameVersion := range matchedConfig.Depedencies {
		if strings.HasPrefix(nameVersion, p.Name) {
			return fmt.Errorf("port.dependencies contains circular dependency: %s", nameVersion)
		}

		packageDir := filepath.Join(Dirs.WorkspaceDir, "packages", nameVersion+"-"+platformProject)
		packageFiles, err := matchedConfig.BuildSystem().PackageFiles(
			packageDir,
			p.ctx.Platform().Name,
			p.ctx.Project().Name,
			p.ctx.BuildType(),
		)
		if err != nil {
			return err
		}

		for _, file := range packageFiles {
			file = strings.TrimPrefix(file, platformProject+"/")
			src := filepath.Join(packageDir, file)
			dest := filepath.Join(matchedConfig.PortConfig.InstalledDir, file)

			if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
				return err
			}
			if err := fileio.CopyFile(src, dest); err != nil {
				return err
			}
		}
	}

	// Check and repair current port.
	packageFiles, err := matchedConfig.BuildSystem().PackageFiles(
		matchedConfig.PortConfig.PackageDir,
		p.ctx.Platform().Name,
		p.ctx.Project().Name,
		p.ctx.BuildType(),
	)
	if err != nil {
		return err
	}

	// No files found, skip it, maybe need to install from source.
	if len(packageFiles) == 0 {
		return nil
	}

	// Copy files from package to installed dir.
	for _, file := range packageFiles {
		if p.AsDev {
			file = strings.TrimPrefix(file, "dev/")
		} else {
			file = strings.TrimPrefix(file, platformProject+"/")
		}

		src := filepath.Join(matchedConfig.PortConfig.PackageDir, file)
		dest := filepath.Join(matchedConfig.PortConfig.InstalledDir, file)

		if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
			return err
		}
		if err := fileio.CopyFile(src, dest); err != nil {
			return err
		}
	}

	return nil
}

func (p Port) downloadAndDeploy(url, installedDir, downloadedDir string) error {
	// Download to fixed dir.
	downloadRequest := fileio.NewDownloadRequest(url, downloadedDir)
	downloaded, err := downloadRequest.Download()
	if err != nil {
		return fmt.Errorf("%s: download port failed: %w", url, err)
	}

	// Extract archive file.
	archiveName := filepath.Base(url)
	folderName := strings.TrimSuffix(archiveName, ".tar.gz")
	extractPath := filepath.Join(installedDir, folderName)
	if err := fileio.Extract(downloaded, extractPath); err != nil {
		return fmt.Errorf("%s: extract %s failed: %w", archiveName, downloaded, err)
	}

	return nil
}

func (p Port) buildCrossTools() buildsystem.CrossTools {
	crossTools := buildsystem.CrossTools{
		SystemName:      p.ctx.SystemName(),
		SystemProcessor: p.ctx.SystemProcessor(),
	}

	if p.ctx.Toolchain() != nil {
		crossTools.FullPath = p.ctx.Toolchain().fullpath
		crossTools.Native = false
		crossTools.Host = p.ctx.Toolchain().Host
		crossTools.ToolchainPrefix = p.ctx.Toolchain().ToolchainPrefix
		crossTools.RootFS = p.ctx.RootFS().fullpath
		crossTools.CC = p.ctx.Toolchain().CC
		crossTools.CXX = p.ctx.Toolchain().CXX
		crossTools.FC = p.ctx.Toolchain().FC
		crossTools.RANLIB = p.ctx.Toolchain().RANLIB
		crossTools.AR = p.ctx.Toolchain().AR
		crossTools.LD = p.ctx.Toolchain().LD
		crossTools.NM = p.ctx.Toolchain().NM
		crossTools.OBJDUMP = p.ctx.Toolchain().OBJDUMP
		crossTools.STRIP = p.ctx.Toolchain().STRIP
	} else {
		crossTools.Native = true
	}

	return crossTools
}

func (p Port) tryRemoveBuildCache(logNamePrefix string) {
	buildTrees := filepath.Join(Dirs.WorkspaceDir, "buildtrees")
	buildCacheDir := filepath.Join(buildTrees, logNamePrefix)
	configureLogPath := filepath.Join(buildTrees, logNamePrefix+"-configure.log")
	buildLogPath := filepath.Join(buildTrees, logNamePrefix+"-build.log")
	installLogPath := filepath.Join(buildTrees, logNamePrefix+"-install.log")

	os.RemoveAll(buildCacheDir)
	os.Remove(configureLogPath)
	os.Remove(buildLogPath)
	os.Remove(installLogPath)
}
