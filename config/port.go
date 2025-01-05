package config

import (
	"buildenv/buildsystem"
	"buildenv/pkg/color"
	"buildenv/pkg/fileio"
	"encoding/json"
	"fmt"
	"io"
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
	fileName := fmt.Sprintf("%s-%s.list", ctx.Platform().Name, ctx.BuildType())

	p.ctx = ctx
	p.installInfoFile = filepath.Join(Dirs.InstalledDir, "buildenv", "info", nameVersion+"-"+fileName)

	// Init build config with rootfs, toolchain info.
	platformBuildType := fmt.Sprintf("%s-%s", ctx.Platform().Name, ctx.BuildType())
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
		SourceDir:       filepath.Join(Dirs.WorkspaceDir, "buildtrees", p.NameVersion(), "src"),
		BuildDir:        filepath.Join(Dirs.WorkspaceDir, "buildtrees", p.NameVersion(), platformBuildType),
		PackageDir:      filepath.Join(Dirs.WorkspaceDir, "packages", p.NameVersion()+"-"+platformBuildType),
		InstalledDir:    filepath.Join(Dirs.InstalledDir, platformBuildType),
	}

	if len(p.BuildConfigs) > 0 {
		for index := range p.BuildConfigs {
			p.BuildConfigs[index].PortConfig = portConfig
		}
	}

	return nil
}

func (p *Port) Verify() error {
	if p.Url == "" {
		return fmt.Errorf("port.url is empty")
	}

	if p.Name == "" {
		return fmt.Errorf("port.name is empty")
	}

	if p.Version == "" {
		return fmt.Errorf("port.version is empty")
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

	// Check if package exists, if exists, install from package,
	// otherwise, install from source.
	if fileio.PathExists(matchedConfig.PortConfig.PackageDir) {
		if err := p.installFromPackage(matchedConfig); err != nil {
			return err
		}
	} else {
		if err := p.installFromSource(silentMode, matchedConfig); err != nil {
			return err
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
		title := color.Sprintf(color.Green, "\n[✔] ---- Port: %s\n", p.NameVersion())
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
	platformBuildType := fmt.Sprintf("%s-%s", p.ctx.Platform().Name, p.ctx.BuildType())

	// First, we must check and repair dependency ports.
	for _, item := range matchedConfig.Depedencies {
		if strings.HasPrefix(item, p.Name) {
			return fmt.Errorf("port.dependencies contains circular dependency: %s", item)
		}

		packageDir := filepath.Join(Dirs.WorkspaceDir, "packages", item+"-"+platformBuildType)
		packageFiles, err := matchedConfig.BuildSystem().PackageFiles(
			packageDir,
			p.ctx.Platform().Name,
			p.ctx.BuildType(),
		)
		if err != nil {
			return err
		}

		for _, file := range packageFiles {
			file = strings.TrimPrefix(file, platformBuildType+"/")
			src := filepath.Join(packageDir, file)
			dest := filepath.Join(matchedConfig.PortConfig.InstalledDir, file)

			if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
				return err
			}
			if err := p.copyFile(src, dest); err != nil {
				return err
			}
		}
	}

	// Check and repair current port.
	packageFiles, err := matchedConfig.BuildSystem().PackageFiles(
		matchedConfig.PortConfig.PackageDir,
		p.ctx.Platform().Name,
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
		file = strings.TrimPrefix(file, platformBuildType+"/")
		src := filepath.Join(matchedConfig.PortConfig.PackageDir, file)
		dest := filepath.Join(matchedConfig.PortConfig.InstalledDir, file)

		if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
			return err
		}
		if err := p.copyFile(src, dest); err != nil {
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

func (p Port) copyFile(src, dest string) error {
	// Read file info.
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}

	// Create symlink if it's a symlink.
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(src)
		if err != nil {
			return err
		}

		// Remove dest if it exists before creating symlink.
		if _, err := os.Lstat(dest); err == nil {
			if removeErr := os.Remove(dest); removeErr != nil {
				return removeErr
			}
		}

		return os.Symlink(target, dest)
	}

	// Copy normal file.
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return err
	}

	return nil
}
