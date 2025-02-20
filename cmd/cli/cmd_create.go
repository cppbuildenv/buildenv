package cli

import (
	"buildenv/config"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func handleCreate(callbacks config.BuildEnvCallbacks) {
	var (
		platform string
		project  string
		tool     string
		port     string
	)

	cmd := flag.NewFlagSet("create", flag.ExitOnError)
	cmd.StringVar(&platform, "platform", "", "")
	cmd.StringVar(&project, "project", "", "")
	cmd.StringVar(&tool, "tool", "", "")
	cmd.StringVar(&port, "port", "", "")

	cmd.Usage = func() {
		fmt.Print("Usage: buildenv create [options]\n\n")
		fmt.Println("options:")
		cmd.PrintDefaults()
	}

	// Check if the create target is specified.
	if len(os.Args) < 2 {
		fmt.Println("Error: one of --platform, --project, --tool and --port must be specified.")
		cmd.Usage()
		os.Exit(1)
	}

	cmd.Parse(os.Args[2:])

	if platform != "" {
		createPlatform(platform)
	} else if project != "" {
		createProject(project)
	} else if tool != "" {
		createTool(tool, callbacks)
	} else if port != "" {
		createPort(port, callbacks)
	}
}

func createPlatform(platformName string) {
	platformName = strings.TrimSpace(platformName)
	platformPath := filepath.Join(config.Dirs.PlatformsDir, platformName)
	if !strings.HasSuffix(platformPath, ".json") {
		platformPath = platformPath + ".json"
	}

	var platform config.Platform
	if err := platform.Write(platformPath); err != nil {
		config.PrintError(err, "%s could not be created.", platformName)
		os.Exit(1)
	}

	config.PrintSuccess("%s is created but need to config it later.", platformName)
}

func createProject(projectName string) {
	projectName = strings.TrimSpace(projectName)
	projectPath := filepath.Join(config.Dirs.ProjectsDir, projectName)
	if !strings.HasSuffix(projectPath, ".json") {
		projectPath = projectPath + ".json"
	}

	var project config.Project
	if err := project.Write(projectPath); err != nil {
		config.PrintSuccess("%s could not be created.", projectName)
		os.Exit(1)
	}

	config.PrintSuccess("%s is created but need to config it later.", projectName)
}

func createTool(toolName string, callbacks config.BuildEnvCallbacks) {
	if err := callbacks.OnCreateTool(toolName); err != nil {
		config.PrintError(err, "%s could not be created.", toolName)
		os.Exit(1)
	}

	config.PrintSuccess(" %s is created but need to config it later.", toolName)
}

func createPort(nameVersion string, callbacks config.BuildEnvCallbacks) {
	if err := callbacks.OnCreatePort(nameVersion); err != nil {
		config.PrintError(err, "%s could not be created.", nameVersion)
		os.Exit(1)
	}

	config.PrintSuccess("%s is created but need to config it later.", nameVersion)
}
