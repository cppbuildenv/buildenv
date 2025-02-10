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
	// Some third-party's configure scripts is not exist in the source folder root.
	b.PortConfig.SourceDir = filepath.Join(b.PortConfig.SourceDir, b.PortConfig.SourceFolder)
	if err := os.Chdir(b.PortConfig.SourceDir); err != nil {
		return err
	}

	// Clean build cache files.
	for _, path := range []string{
		"b2",
		"b2.exe",
		"bin.v2",
		"project-config.jam",
		"tools/build/src/engine/b2",
	} {
		os.RemoveAll(path)
	}

	b.setBuildType(buildType)

	// Append common variables for cross compiling.
	b.Arguments = append(b.Arguments, fmt.Sprintf("--prefix=%s", b.PortConfig.PackageDir))

	// Join args into a string.
	joinedArgs := strings.Join(b.Arguments, " ")
	configure := fmt.Sprintf("./bootstrap.sh %s", joinedArgs)

	// Execute configure.
	logPath := b.getLogPath("configure")
	title := fmt.Sprintf("[configure %s@%s]", b.PortConfig.LibName, b.PortConfig.LibVersion)
	executor := cmd.NewExecutor(title, configure)
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
	if err := os.Chdir(b.PortConfig.SourceDir); err != nil {
		return err
	}

	rootfs := b.PortConfig.CrossTools.RootFS
	b.Arguments = append(b.Arguments, "toolset=gcc")
	if !b.AsDev {
		b.Arguments = append(b.Arguments, "cxxflags=--sysroot=%s", rootfs)
		b.Arguments = append(b.Arguments, "linkflags=--sysroot=%s", rootfs)
	}

	b.prepareBuildInstall()

	// Assemble command.
	joinedArgs := strings.Join(b.Arguments, " ")
	command := fmt.Sprintf("./b2 %s -j %d", joinedArgs, b.PortConfig.JobNum)

	// Execute build.
	logPath := b.getLogPath("build")
	title := fmt.Sprintf("[build %s@%s]", b.PortConfig.LibName, b.PortConfig.LibVersion)
	executor := cmd.NewExecutor(title, command)
	executor.SetLogPath(logPath)
	if err := executor.Execute(); err != nil {
		return err
	}

	return nil
}

func (b b2) Install() error {
	b.prepareBuildInstall()

	// Assemble command.
	joinedArgs := strings.Join(b.Arguments, " ")
	command := fmt.Sprintf("./b2 install %s", joinedArgs)

	// Execute install.
	logPath := b.getLogPath("install")
	title := fmt.Sprintf("[install %s@%s]", b.PortConfig.LibName, b.PortConfig.LibVersion)
	executor := cmd.NewExecutor(title, command)
	executor.SetLogPath(logPath)
	if err := executor.Execute(); err != nil {
		return err
	}

	return nil
}

func (b b2) prepareBuildInstall() {
	// "--with-libraries" and "--without-libraries" should be removed during build and install.
	b.Arguments = slices.DeleteFunc(b.Arguments, func(element string) bool {
		return strings.HasPrefix(element, "--with-libraries") ||
			strings.HasPrefix(element, "--without-libraries")
	})

	// Override library type if specified.
	if b.BuildConfig.LibraryType != "" {
		b.Arguments = slices.DeleteFunc(b.Arguments, func(element string) bool {
			return strings.HasPrefix(element, "link=") ||
				strings.HasPrefix(element, "runtime-link=")
		})

		switch b.BuildConfig.LibraryType {
		case "static":
			b.Arguments = append(b.Arguments, "link=static")
			b.Arguments = append(b.Arguments, "runtime-link=static")

		case "shared":
			b.Arguments = append(b.Arguments, "link=shared")
			b.Arguments = append(b.Arguments, "runtime-link=shared")
		}
	}
}
