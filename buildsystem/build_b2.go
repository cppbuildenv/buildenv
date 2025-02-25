package buildsystem

import (
	"bufio"
	"buildenv/pkg/cmd"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func NewB2(config BuildConfig) *b2 {
	return &b2{BuildConfig: config}
}

type b2 struct {
	BuildConfig
}

func (b *b2) Configure(buildType string) error {
	// Some libraries' configure or CMakeLists.txt may not in root folder.
	b.PortConfig.SourceDir = filepath.Join(b.PortConfig.SourceDir, b.PortConfig.SourceFolder)

	b.setBuildType(buildType)

	// Append common options for cross compiling.
	b.Options = append(b.Options, fmt.Sprintf("--prefix=%s", b.PortConfig.PackageDir))

	// Join options into a string.
	joinedArgs := strings.Join(b.Options, " ")
	configure := fmt.Sprintf("%s/bootstrap.sh %s", b.PortConfig.SourceDir, joinedArgs)

	// Execute configure.
	logPath := b.getLogPath("configure")
	title := fmt.Sprintf("[configure %s@%s]", b.PortConfig.LibName, b.PortConfig.LibVersion)
	executor := cmd.NewExecutor(title, configure)
	executor.SetWorkDir(b.PortConfig.SourceDir)
	executor.SetLogPath(logPath)
	if err := executor.Execute(); err != nil {
		return err
	}

	// Modify project-config.jam to set cross-compiling tool for none-runtime library.
	if !b.AsDev {
		configPath := filepath.Join(b.PortConfig.SourceDir, "project-config.jam")
		file, err := os.OpenFile(configPath, os.O_RDONLY, 0755)
		if err != nil {
			return err
		}
		defer file.Close()

		// Override project-config.jam.
		var buffer bytes.Buffer
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "using gcc ;") {
				line = fmt.Sprintf("    using gcc : : %sg++ ;", b.PortConfig.CrossTools.ToolchainPrefix)
				buffer.WriteString(line + "\n")
			} else {
				buffer.WriteString(line + "\n")
			}
		}

		// Write override `project-config.jam`.
		if err := os.WriteFile(configPath, buffer.Bytes(), 0755); err != nil {
			return err
		}
	}

	return nil
}

func (b b2) Build() error {
	b.Options = append(b.Options, fmt.Sprintf("--build-dir=%s", b.PortConfig.BuildDir))
	b.Options = append(b.Options, "toolset=gcc")

	if !b.AsDev {
		rootfs := b.PortConfig.CrossTools.RootFS
		b.Options = append(b.Options, "cxxflags=--sysroot=%s", rootfs)
		b.Options = append(b.Options, "linkflags=--sysroot=%s", rootfs)
	}

	b.prepareBuildInstall()

	// Assemble command.
	joinedArgs := strings.Join(b.Options, " ")
	command := fmt.Sprintf("%s/b2 %s -j %d", b.PortConfig.SourceDir, joinedArgs, b.PortConfig.JobNum)

	// Execute build.
	logPath := b.getLogPath("build")
	title := fmt.Sprintf("[build %s@%s]", b.PortConfig.LibName, b.PortConfig.LibVersion)
	executor := cmd.NewExecutor(title, command)
	executor.SetWorkDir(b.PortConfig.SourceDir)
	executor.SetLogPath(logPath)
	if err := executor.Execute(); err != nil {
		return err
	}

	return nil
}

func (b b2) Install() error {
	b.prepareBuildInstall()

	// Assemble command.
	joinedOptions := strings.Join(b.Options, " ")
	command := fmt.Sprintf("%s/b2 install %s", b.PortConfig.SourceDir, joinedOptions)

	// Execute install.
	logPath := b.getLogPath("install")
	title := fmt.Sprintf("[install %s@%s]", b.PortConfig.LibName, b.PortConfig.LibVersion)
	executor := cmd.NewExecutor(title, command)
	executor.SetWorkDir(b.PortConfig.SourceDir)
	executor.SetLogPath(logPath)
	if err := executor.Execute(); err != nil {
		return err
	}

	return nil
}

func (b b2) prepareBuildInstall() {
	// "--with-libraries" and "--without-libraries" should be removed during build and install.
	b.Options = slices.DeleteFunc(b.Options, func(element string) bool {
		return strings.HasPrefix(element, "--with-libraries") ||
			strings.HasPrefix(element, "--without-libraries")
	})

	// Override library type if specified.
	if b.BuildConfig.LibraryType != "" {
		b.Options = slices.DeleteFunc(b.Options, func(element string) bool {
			return strings.HasPrefix(element, "link=") ||
				strings.HasPrefix(element, "runtime-link=")
		})

		switch b.BuildConfig.LibraryType {
		case "static":
			b.Options = append(b.Options, "link=static")
			b.Options = append(b.Options, "runtime-link=static")

		case "shared":
			b.Options = append(b.Options, "link=shared")
			b.Options = append(b.Options, "runtime-link=shared")
		}
	}
}
