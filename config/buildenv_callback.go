package config

import (
	"buildenv/pkg/color"
	"buildenv/pkg/io"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var Callbacks = callbackImpl{}

type callbackImpl struct{}

func (c callbackImpl) OnInitBuildEnv(confRepoUrl, confRepoRef string) (string, error) {
	buildenv := NewBuildEnv()

	// Create buildenv.json if not exist.
	confPath := filepath.Join(Dirs.WorkspaceDir, "buildenv.json")
	if !io.PathExists(confPath) {
		if err := os.MkdirAll(filepath.Dir(confPath), os.ModePerm); err != nil {
			return "", err
		}

		buildenv.ConfRepoUrl = confRepoUrl
		buildenv.ConfRepoRef = confRepoRef
		bytes, err := json.MarshalIndent(buildenv, "", "    ")
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(confPath, []byte(bytes), os.ModePerm); err != nil {
			return "", err
		}
	}

	// Sync conf repo with repo url.
	bytes, err := os.ReadFile(confPath)
	if err != nil {
		return "", err
	}

	// Unmarshall with buildenv.json.
	if err := json.Unmarshal(bytes, &buildenv); err != nil {
		return "", err
	}

	// Override buildenv.json with repo url and repo ref.
	if confRepoUrl != "" && confRepoRef != "" {
		buildenv.ConfRepoUrl = confRepoUrl
		buildenv.ConfRepoRef = confRepoRef

		if err := os.WriteFile(confPath, []byte(bytes), os.ModePerm); err != nil {
			return "", err
		}
	}

	// Sync repo.
	return buildenv.Synchronize(buildenv.ConfRepoUrl, buildenv.ConfRepoRef)
}

func (c callbackImpl) OnCreatePlatform(platformName string) error {
	if platformName == "" {
		return fmt.Errorf("platformName is empty for creating new platform")
	}

	// Create platform file.
	platformPath := filepath.Join(Dirs.PlatformsDir, platformName+".json")
	var platform Platform
	if err := platform.Write(platformPath); err != nil {
		return err
	}

	return nil
}

func (c callbackImpl) OnSelectPlatform(platformName string) error {
	// Init buildenv with "buildenv.json"
	buildenv := NewBuildEnv()
	buildEnvPath := filepath.Join(Dirs.WorkspaceDir, "buildenv.json")

	if !io.PathExists(buildEnvPath) { // Create buildenv.json if not exist.
		// Create conf directory.
		if err := os.MkdirAll(filepath.Dir(buildEnvPath), os.ModeDir|os.ModePerm); err != nil {
			return err
		}

		// Create buildenv conf file with default values.
		buildenv.configData.JobNum = runtime.NumCPU()
		bytes, err := json.MarshalIndent(buildenv, "", "    ")
		if err != nil {
			return fmt.Errorf("cannot marshal buildenv conf: %w", err)
		}
		if err := os.WriteFile(buildEnvPath, bytes, os.ModePerm); err != nil {
			return err
		}
	} else { // Read buildenv.json.
		bytes, err := os.ReadFile(buildEnvPath)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(bytes, &buildenv); err != nil {
			return fmt.Errorf("cannot unmarshal buildenv conf: %w", err)
		}
	}

	// Init and verify platform.
	args := NewVerifyArgs(false, false)
	if err := buildenv.platform.Init(buildenv, platformName); err != nil {
		return err
	}
	buildenv.configData.PlatformName = platformName
	if err := buildenv.platform.Verify(args); err != nil {
		return err
	}

	// Do change platform.
	bytes, err := json.MarshalIndent(buildenv, "", "    ")
	if err != nil {
		return fmt.Errorf("cannot marshal buildenv conf: %w", err)
	}
	if err := os.WriteFile(buildEnvPath, bytes, os.ModePerm); err != nil {
		return err
	}

	return nil
}

func (c callbackImpl) OnCreateProject(projectName string) error {
	if projectName == "" {
		return fmt.Errorf("projectName is empty for creating new project")
	}

	// Create project file.
	projectPath := filepath.Join(Dirs.ProjectsDir, projectName+".json")
	var project Project
	if err := project.Write(projectPath); err != nil {
		return err
	}

	return nil
}

func (c callbackImpl) OnSelectProject(projectName string) error {
	// Init buildenv with "buildenv.json"
	buildenv := NewBuildEnv()
	buildEnvPath := filepath.Join(Dirs.WorkspaceDir, "buildenv.json")

	if !io.PathExists(buildEnvPath) { // Create buildenv.json if not exist.
		// Create conf directory.
		if err := os.MkdirAll(filepath.Dir(buildEnvPath), os.ModeDir|os.ModePerm); err != nil {
			return err
		}

		// Create buildenv conf file with default values.
		buildenv.configData.JobNum = runtime.NumCPU()
		bytes, err := json.MarshalIndent(buildenv, "", "    ")
		if err != nil {
			return fmt.Errorf("cannot marshal buildenv conf: %w", err)
		}
		if err := os.WriteFile(buildEnvPath, bytes, os.ModePerm); err != nil {
			return err
		}

		return fmt.Errorf("please select a platform first")
	}

	// Read buildenv.json to check if platform is selected.
	bytes, err := os.ReadFile(buildEnvPath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bytes, &buildenv); err != nil {
		return fmt.Errorf("cannot unmarshal buildenv conf: %w", err)
	}
	if buildenv.PlatformName == "" {
		return fmt.Errorf("please select a platform first")
	}

	args := NewVerifyArgs(false, false)

	// Read platform file and verify it.
	if err := buildenv.platform.Init(buildenv, buildenv.PlatformName); err != nil {
		return err
	}
	if err := buildenv.platform.Verify(args); err != nil {
		return err
	}

	// Read project file and verify it.
	if err := buildenv.project.Init(buildenv, projectName); err != nil {
		return err
	}
	buildenv.ProjectName = projectName
	if err := buildenv.project.Verify(args); err != nil {
		return err
	}

	// Do change project.
	bytes, err = json.MarshalIndent(buildenv, "", "    ")
	if err != nil {
		return fmt.Errorf("cannot marshal buildenv conf: %w", err)
	}
	if err := os.WriteFile(buildEnvPath, bytes, os.ModePerm); err != nil {
		return err
	}

	// Generate toolchain file.
	scriptDir := filepath.Join(Dirs.WorkspaceDir, "script")
	if _, err := buildenv.GenerateToolchainFile(scriptDir); err != nil {
		return err
	}

	return nil
}

func (c callbackImpl) OnCreateTool(toolName string) error {
	toolPath := filepath.Join(Dirs.ToolsDir, toolName+".json")

	var tool Tool
	if err := tool.Write(toolPath); err != nil {
		return err
	}

	return nil
}

func (c callbackImpl) OnCreatePort(portNameVersion string) error {
	parts := strings.Split(portNameVersion, "@")
	if len(parts) != 2 {
		return fmt.Errorf("invalid port name version")
	}

	parentDir := filepath.Join(Dirs.PortsDir, parts[0])
	if err := os.MkdirAll(parentDir, os.ModePerm); err != nil {
		return err
	}
	portPath := filepath.Join(parentDir, parts[1]+".json")

	var port Port
	if err := port.Write(portPath); err != nil {
		return err
	}

	return nil
}

func (c callbackImpl) About(version string) string {
	toolchainPath, _ := filepath.Abs("script/toolchain_file.cmake")
	environmentPath, _ := filepath.Abs("script/environment")
	environmentPath = color.Sprintf(color.Magenta, "%s", environmentPath)
	toolchainPath = color.Sprintf(color.Magenta, "%s", toolchainPath)

	return fmt.Sprintf("\nWelcome to buildenv (%s).\n"+
		"---------------------------------------\n"+
		"This is a simple pkg-manager for C/C++.\n\n"+
		"1. How to use it to build cmake project: \n"+
		"option1: %s\n"+
		"option2: %s\n\n"+
		"2. How to use it to build makefile project: \n"+
		"%s\n\n"+
		"%s",
		version,
		color.Sprintf(color.Blue, "set(CMAKE_TOOLCHAIN_FILE \"%s\")", toolchainPath),
		color.Sprintf(color.Blue, "cmake .. -DCMAKE_TOOLCHAIN_FILE=%s", toolchainPath),
		color.Sprintf(color.Blue, "source %s", environmentPath),
		color.Sprintf(color.Gray, "[ctrl+c/q -> quit]"),
	)
}
