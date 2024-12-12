package config

import (
	"buildenv/pkg/color"
	"fmt"
	"path/filepath"
)

var Callbacks = callbackImpl{}

type callbackImpl struct{}

func (c callbackImpl) OnCreatePlatform(platformName string) error {
	if platformName == "" {
		return fmt.Errorf("platformName is empty for creating new platform")
	}

	// Create platform file.
	platformPath := filepath.Join(Dirs.PlatformDir, platformName+".json")
	var platform Platform
	if err := platform.Write(platformPath); err != nil {
		return err
	}

	return nil
}

func (c callbackImpl) OnSelectPlatform(platformName string) error {
	// In config mode, we always regard build type as `Release`.
	buildType := "Release"

	buildenv := NewBuildEnv(buildType)
	if err := buildenv.ChangePlatform(platformName); err != nil {
		return err
	}

	// Verify buildenv to check if all the required fields are set for generate toolchain file.
	if err := buildenv.Verify(NewVerifyArgs(false, false, buildType)); err != nil {
		return err
	}

	// Generate toolchain file.
	scriptDir := filepath.Join(Dirs.WorkspaceDir, "script")
	if _, err := buildenv.platform.GenerateToolchainFile(scriptDir); err != nil {
		return err
	}

	return nil
}

func (c callbackImpl) OnCreateProject(projectName string) error {
	if projectName == "" {
		return fmt.Errorf("projectName is empty for creating new project")
	}

	// Create project file.
	projectPath := filepath.Join(Dirs.ProjectDir, projectName+".json")
	var project Project
	if err := project.Write(projectPath); err != nil {
		return err
	}

	return nil
}

func (c callbackImpl) OnSelectProject(projectName string) error {
	// In config mode, we always regard build type as `Release`.
	buildType := "Release"

	buildenv := NewBuildEnv(buildType)
	if err := buildenv.ChangeProject(projectName); err != nil {
		return err
	}

	// Verify buildenv to check if all the required fields are set for generate toolchain file.
	if err := buildenv.Verify(NewVerifyArgs(false, false, buildType)); err != nil {
		return err
	}

	// Generate toolchain file.
	scriptDir := filepath.Join(Dirs.WorkspaceDir, "script")
	if _, err := buildenv.platform.GenerateToolchainFile(scriptDir); err != nil {
		return err
	}

	return nil
}

func (c callbackImpl) About(version string) string {
	toolchainPath, _ := filepath.Abs("script/toolchain_file.cmake")
	environmentPath, _ := filepath.Abs("script/environment")

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
