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

type BuildTool int

type Port struct {
	Url          string                    `json:"url"`
	Name         string                    `json:"name"`
	Version      string                    `json:"version"`
	SourceFolder string                    `json:"source_folder,omitempty"`
	BuildConfigs []buildsystem.BuildConfig `json:"build_configs"`

	// Internal fields.
	ctx             Context
	installInfoFile string // used to record installed state
	isSubDep        bool
}

func (p Port) NameVersion() string {
	return p.Name + "@" + p.Version
}

func (p *Port) Init(ctx Context, portPath string) error {
	portPath = strings.ReplaceAll(portPath, "@", "/")
	if !fileio.PathExists(portPath) {
		portName := fileio.FileBaseName(portPath)
		if p.isSubDep {
			return fmt.Errorf("sub depedency port %s does not exists", portName)
		} else {
			return fmt.Errorf("port %s does not exists", portName)
		}
	}

	bytes, err := os.ReadFile(portPath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, p); err != nil {
		return err
	}

	// Info file: used to record installed state.
	nameVersion := p.NameVersion()
	platformProject := fmt.Sprintf("%s@%s@%s.list", ctx.Platform().Name, ctx.Project().Name, ctx.BuildType())

	p.ctx = ctx
	p.installInfoFile = filepath.Join(Dirs.InstalledDir, "buildenv", "info", nameVersion+"@"+platformProject)

	// Init build config with rootfs, toolchain info.
	buildFolder := filepath.Join(p.NameVersion(), ctx.Platform().Name+"@"+ctx.Project().Name+"@"+ctx.BuildType())
	packageFolder := fmt.Sprintf("%s@%s@%s@%s", p.NameVersion(), ctx.Platform().Name, ctx.Project().Name, ctx.BuildType())
	installedFolder := fmt.Sprintf("%s@%s@%s", ctx.Platform().Name, ctx.Project().Name, ctx.BuildType())

	portConfig := buildsystem.PortConfig{
		SystemName:      ctx.SystemName(),
		SystemProcessor: ctx.SystemProcessor(),
		Host:            ctx.Host(),
		RootFS:          ctx.RootFSPath(),
		ToolchainPrefix: ctx.ToolchainPrefix(),
		JobNum:          ctx.JobNum(),
		LibName:         p.Name,
		LibVersion:      p.Version,
		SourceFolder:    p.SourceFolder,
		PortsDir:        Dirs.PortsDir,
		DownloadedDir:   Dirs.DownloadedDir,
		SourceDir:       filepath.Join(Dirs.WorkspaceDir, "buildtrees", p.NameVersion(), "src"),
		BuildDir:        filepath.Join(Dirs.WorkspaceDir, "buildtrees", buildFolder),
		PackageDir:      filepath.Join(Dirs.WorkspaceDir, "packages", packageFolder),
		InstalledDir:    filepath.Join(Dirs.InstalledDir, installedFolder),
		TmpDir:          filepath.Join(Dirs.WorkspaceDir, "tmp"),
	}

	if len(p.BuildConfigs) > 0 {
		for index := range p.BuildConfigs {
			p.BuildConfigs[index].PortConfig = portConfig

			// Merge project override ports.
			p.mergeBuildConfig(&p.BuildConfigs[index], ctx.Project().OverridePorts)
		}
	}

	return nil
}

func (p *Port) Verify() error {
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

		if err := config.Verify(); err != nil {
			return err
		}
	}

	return nil
}

func (p Port) Installed() bool {
	if !fileio.PathExists(p.installInfoFile) {
		return false
	}

	// File can be read?
	bytes, err := os.ReadFile(p.installInfoFile)
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
	installedDir := filepath.Join(Dirs.WorkspaceDir, "installed", p.ctx.Platform().Name+"-"+p.ctx.BuildType())
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
			platformProject := fmt.Sprintf("%s@%s@%s", p.ctx.Platform().Name, p.ctx.Project().Name, p.ctx.BuildType())
			logPathPrefix := filepath.Join(p.NameVersion(), platformProject)
			p.tryRemoveBuildCache(logPathPrefix)

			// Install from source when cache not found.
			if err := p.installFromSource(silentMode, matchedConfig); err != nil {
				return err
			}

			// Write package to cache dirs so that others can share installed libraries.
			for _, cacheDir := range p.ctx.CacheDirs() {
				if !cacheDir.Writable {
					continue
				}

				if err := cacheDir.Write(matchedConfig.PortConfig.PackageDir); err != nil {
					return err
				}
			}

			installedFrom = "source"
		}

		// This will copy all install files into installedDir.
		if err := p.installFromPackage(matchedConfig); err != nil {
			return err
		}
	}

	// Write installed file list info into its installation info list.
	if err := os.MkdirAll(filepath.Dir(p.installInfoFile), os.ModePerm); err != nil {
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
	if err := os.WriteFile(p.installInfoFile, []byte(strings.Join(packageFiles, "\n")), os.ModePerm); err != nil {
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
	}
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

func (p Port) installFromSource(silentMode bool, matchedConfig *buildsystem.BuildConfig) error {
	// First, we must check and repair dependency ports.
	for _, item := range matchedConfig.Depedencies {
		if strings.HasPrefix(item, p.Name) {
			return fmt.Errorf("port.dependencies contains circular dependency: %s", item)
		}

		// Check and repair dependency.
		var port Port
		port.isSubDep = true
		portPath := filepath.Join(Dirs.PortsDir, item+".json")
		if err := port.Init(p.ctx, portPath); err != nil {
			return err
		}
		if err := port.Verify(); err != nil {
			return err
		}
		if err := port.Install(silentMode); err != nil {
			return err
		}
	}

	// Check and repair current port.
	if err := matchedConfig.Install(p.Url, p.Version, p.ctx.BuildType()); err != nil {
		return err
	}

	return nil
}

func (p Port) installFromPackage(matchedConfig *buildsystem.BuildConfig) error {
	platformProject := fmt.Sprintf("%s@%s@%s", p.ctx.Platform().Name, p.ctx.Project().Name, p.ctx.BuildType())

	// First, we must check and repair dependency ports.
	for _, nameVersion := range matchedConfig.Depedencies {
		if strings.HasPrefix(nameVersion, p.Name) {
			return fmt.Errorf("port.dependencies contains circular dependency: %s", nameVersion)
		}

		packageDir := filepath.Join(Dirs.WorkspaceDir, "packages", nameVersion+"@"+platformProject)
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
		file = strings.TrimPrefix(file, platformProject+"/")
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
