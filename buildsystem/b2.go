package buildsystem

import (
	"bufio"
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
	// Clean repo source before configuration.
	if err := cleanRepo(b.PortConfig.SourceDir); err != nil {
		return err
	}

	if err := os.Chdir(b.PortConfig.SourceDir); err != nil {
		return err
	}

	// Append common variables for cross compiling.
	b.Arguments = append(b.Arguments, fmt.Sprintf("--prefix=%s", b.PortConfig.PackageDir))

	// Join args into a string.
	joinedArgs := strings.Join(b.Arguments, " ")
	configure := fmt.Sprintf("./bootstrap.sh %s", joinedArgs)

	// Execute configure.
	parentDir := filepath.Dir(b.PortConfig.BuildDir)
	fileName := filepath.Base(b.PortConfig.BuildDir) + "-configure.log"
	configureLogPath := filepath.Join(parentDir, fileName)
	title := fmt.Sprintf("[configure %s]", b.PortConfig.LibName)
	if err := execute(title, configure, configureLogPath); err != nil {
		return err
	}

	// Modify project-config.jam to set cross-compiling tool.
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
			line = fmt.Sprintf("    using gcc : : %sg++ ;", b.PortConfig.ToolchainPrefix)
			buffer.WriteString(line + "\n")
		} else {
			buffer.WriteString(line + "\n")
		}
	}

	if err := os.WriteFile(configPath, buffer.Bytes(), 0755); err != nil {
		return err
	}

	return nil
}

func (b b2) Build() error {
	if err := os.Chdir(b.PortConfig.SourceDir); err != nil {
		return err
	}

	b.Arguments = append(b.Arguments, "toolset=gcc")
	b.Arguments = append(b.Arguments, "cxxflags=--sysroot=%s", b.PortConfig.RootFS)
	b.Arguments = append(b.Arguments, "linkflags=--sysroot=%s", b.PortConfig.RootFS)

	b.adjustForBuildInstall()

	// Assemble script.
	joinedArgs := strings.Join(b.Arguments, " ")
	command := fmt.Sprintf("./b2 %s -j %d", joinedArgs, b.PortConfig.JobNum)

	// Execute build.
	parentDir := filepath.Dir(b.PortConfig.BuildDir)
	fileName := filepath.Base(b.PortConfig.BuildDir) + "-build.log"
	buildLogPath := filepath.Join(parentDir, fileName)
	title := fmt.Sprintf("[build %s]", b.PortConfig.LibName)
	if err := execute(title, command, buildLogPath); err != nil {
		return err
	}

	return nil
}

func (b b2) Install() error {
	b.adjustForBuildInstall()

	// Assemble script.
	joinedArgs := strings.Join(b.Arguments, " ")
	command := fmt.Sprintf("./b2 install %s", joinedArgs)

	// Execute install.
	parentDir := filepath.Dir(b.PortConfig.BuildDir)
	fileName := filepath.Base(b.PortConfig.BuildDir) + "-install.log"
	installLogPath := filepath.Join(parentDir, fileName)
	title := fmt.Sprintf("[install %s]", b.PortConfig.LibName)
	if err := execute(title, command, installLogPath); err != nil {
		return err
	}

	return nil
}

func (b b2) adjustForBuildInstall() {
	// During build and install, we don't need "--with-libraries" and "--without-libraries".
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
