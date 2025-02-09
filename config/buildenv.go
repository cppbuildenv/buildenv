package config

import (
	"buildenv/pkg/env"
	"buildenv/pkg/fileio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Context interface {
	BuildEnvDir() string
	Platform() Platform
	Project() Project
	Toolchain() *Toolchain
	RootFS() *RootFS
	BuildType() string
	JobNum() int
	CacheDirs() []CacheDir
	SystemName() string
	SystemProcessor() string
	Host() string
	ToolchainPrefix() string
	RootFSPath() string
}

func NewBuildEnv() *buildenv {
	return &buildenv{
		configData: configData{
			JobNum:    runtime.NumCPU(),
			CacheDirs: []CacheDir{},
		},
		buildType: "Release",
	}
}

type buildenv struct {
	configData

	// Internal fields.
	platform  Platform
	project   Project
	buildType string
}

type configData struct {
	ConfRepoUrl  string     `json:"conf_repo_url"`
	ConfRepoRef  string     `json:"conf_repo_ref"`
	PlatformName string     `json:"platform_name"`
	ProjectName  string     `json:"project_name"`
	JobNum       int        `json:"job_num"`
	CacheDirs    []CacheDir `json:"cache_dirs"`
}

func (b *buildenv) SetBuildType(buildType string) *buildenv {
	if buildType == "" {
		buildType = "Release"
	}

	b.buildType = buildType
	return b
}

func (b *buildenv) Setup(args SetupArgs) error {
	buildEnvPath := filepath.Join(Dirs.WorkspaceDir, "buildenv.json")
	if err := b.Init(buildEnvPath); err != nil {
		return err
	}

	// init and setup platform.
	if err := b.platform.Init(b, b.PlatformName); err != nil {
		return err
	}
	if err := b.platform.Setup(args); err != nil {
		return err
	}

	// Append runtime bin path to PATH, this is required by some third-party libraries during build.
	os.Setenv("PATH", filepath.Join(Dirs.InstalledDir, "dev", "bin")+string(os.PathListSeparator)+os.Getenv("PATH"))

	// init and setup project.
	if err := b.project.Init(b, b.ProjectName); err != nil {
		return err
	}
	if err := b.project.Setup(args); err != nil {
		return err
	}

	return nil
}

func (b buildenv) SyncRepo(repo, ref string) (string, error) {
	if b.ConfRepoUrl == "" {
		return "", fmt.Errorf("no conf repo has been provided for buildenv")
	}

	if b.ConfRepoRef == "" {
		return "", fmt.Errorf("no conf repo ref has been provided for buildenv")
	}

	var commands []string

	// Clone or git checkout repo.
	confDir := filepath.Join(Dirs.WorkspaceDir, "conf")
	if fileio.PathExists(confDir) {
		// clean up and checkout to ref.
		if fileio.PathExists(filepath.Join(confDir, ".git")) {
			// cd `conf`` to execute git command.
			if err := os.Chdir(confDir); err != nil {
				return "", err
			}

			commands = append(commands, "git reset --hard && git clean -xfd")
			commands = append(commands, fmt.Sprintf("git -C %s fetch", confDir))
			commands = append(commands, fmt.Sprintf("git -C %s checkout %s", confDir, ref))
			commands = append(commands, "git pull")
		} else {
			// clean up and clone.
			commands = append(commands, fmt.Sprintf("rm -rf %s", confDir))
			commands = append(commands, fmt.Sprintf("git clone --branch %s --single-branch %s %s", ref, repo, confDir))
		}
	} else {
		commands = append(commands, fmt.Sprintf("git clone --branch %s --single-branch %s %s", ref, repo, confDir))
	}

	// Execute clone command.
	commandLine := strings.Join(commands, " && ")
	output, err := b.execute(commandLine, Dirs.WorkspaceDir)
	if err != nil {
		return "", err
	}

	return output, nil
}

func (b buildenv) GenerateToolchainFile(scriptsDir string) (string, error) {
	var toolchain, environment strings.Builder

	// Setup buildenv during configuration.
	toolchain.WriteString(`# This is generated by buildenv. (Do not change it manually!)

# Set default CMAKE_BUILD_TYPE.
if(NOT CMAKE_BUILD_TYPE)
	set(CMAKE_BUILD_TYPE "Release")
endif()

# Setup buildenv during configuration.
set(HOME_DIR "${CMAKE_CURRENT_LIST_DIR}/..")
find_program(BUILDENV buildenv PATHS ${HOME_DIR})
if(BUILDENV)
	execute_process(
		COMMAND ${BUILDENV} -setup -silent -build_type=${CMAKE_BUILD_TYPE}
		WORKING_DIRECTORY ${HOME_DIR}
	)
endif()` + "\n")

	// Define buildenv root dir.
	toolchain.WriteString(fmt.Sprintf("\n%s\n", `# Define buildenv root dir.
get_filename_component(_CURRENT_DIR "${CMAKE_CURRENT_LIST_FILE}" PATH)
get_filename_component(BUILDENV_ROOT_DIR "${_CURRENT_DIR}" PATH)`))

	environment.WriteString("# This is generated by buildenv. (Do not change it manually!)\n")
	environment.WriteString("\n# Define buildenv root dir.\n")
	environment.WriteString("export BUILDENV_ROOT_DIR=$(dirname \"$(dirname \"$BASH_SOURCE\")\")\n")

	// Set sysroot for cross-compile.
	if b.RootFS() != nil {
		if err := b.RootFS().generate(&toolchain, &environment); err != nil {
			return "", err
		}
	}

	// Set toolchain for cross-compile.
	if b.Toolchain() != nil {
		// Set toolchain platform infos.
		toolchain.WriteString("\n# Set toolchain platform infos.\n")
		toolchain.WriteString(fmt.Sprintf("set(CMAKE_SYSTEM_NAME \"%s\")\n", b.SystemName()))
		toolchain.WriteString(fmt.Sprintf("set(CMAKE_SYSTEM_PROCESSOR \"%s\")\n", b.SystemProcessor()))

		if err := b.platform.Toolchain.generate(&toolchain, &environment); err != nil {
			return "", err
		}
	}

	// Set tools for cross-compile.
	if err := b.writeTools(&toolchain, &environment); err != nil {
		return "", err
	}

	toolchain.WriteString("\n# Add `installed dir` into library search paths.\n")
	platformProject := fmt.Sprintf("%s^%s^${CMAKE_BUILD_TYPE}", b.PlatformName, b.ProjectName)
	installedDir := fmt.Sprintf("${BUILDENV_ROOT_DIR}/installed/%s", platformProject)
	toolchain.WriteString(fmt.Sprintf("list(APPEND CMAKE_FIND_ROOT_PATH \"%s\")\n", installedDir))
	toolchain.WriteString(fmt.Sprintf("list(APPEND CMAKE_PREFIX_PATH \"%s\")\n", installedDir))
	toolchain.WriteString(fmt.Sprintf("set(ENV{PKG_CONFIG_PATH} \"%s/lib/pkgconfig%s$ENV{PKG_CONFIG_PATH}\")\n",
		installedDir, string(os.PathListSeparator)))

	// Define cmake vars, env vars and micro vars for project.
	for index, item := range b.project.CMakeVars {
		if index == 0 {
			toolchain.WriteString("\n# Define cmake vars for project.\n")
		}

		parts := strings.Split(item, "=")
		if len(parts) == 1 {
			toolchain.WriteString(fmt.Sprintf("set(%s CACHE INTERNAL \"defined by buildenv globally.\")\n", item))
		} else if len(parts) == 2 {
			toolchain.WriteString(fmt.Sprintf("set(%s \"%s\" CACHE INTERNAL \"defined by buildenv globally.\")\n", parts[0], parts[1]))
		} else {
			return "", fmt.Errorf("invalid cmake var: %s", item)
		}
	}
	for index, item := range b.project.EnvVars {
		parts := strings.Split(item, "=")
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid env var: %s", item)
		}

		if index == 0 {
			toolchain.WriteString("\n# Define env vars for project.\n")
			environment.WriteString("\n# Define env vars for project.\n")
		}
		toolchain.WriteString(fmt.Sprintf("set (ENV{%s} \"%s\")\n", parts[0], parts[1]))
		environment.WriteString(fmt.Sprintf("export %s=%s\n", parts[0], parts[1]))
	}
	for index, item := range b.project.MicroVars {
		if index == 0 {
			toolchain.WriteString("\n# Define micro vars for project.\n")
		}
		toolchain.WriteString(fmt.Sprintf("add_compile_definitions(%s)\n", item))
	}

	// Create the output directory if it doesn't exist.
	if err := os.MkdirAll(scriptsDir, os.ModeDir|os.ModePerm); err != nil {
		return "", err
	}

	// Write toolchain file.
	toolchainPath := filepath.Join(scriptsDir, "toolchain_file.cmake")
	if err := os.WriteFile(toolchainPath, []byte(toolchain.String()), os.ModePerm); err != nil {
		return "", err
	}

	// Write environment file.
	environmentPath := filepath.Join(scriptsDir, "environment")
	if err := os.WriteFile(environmentPath, []byte(environment.String()), os.ModePerm); err != nil {
		return "", err
	}

	// Grant executable permission to the file: rwxr-xr-x
	if err := os.Chmod(environmentPath, 0755); err != nil {
		log.Fatalf("Error setting permissions: %v", err)
	}

	return toolchainPath, nil
}

func (b *buildenv) Init(buildEnvPath string) error {
	if !fileio.PathExists(buildEnvPath) {
		// Create conf directory.
		if err := os.MkdirAll(filepath.Dir(buildEnvPath), os.ModeDir|os.ModePerm); err != nil {
			return err
		}

		b.configData.JobNum = runtime.NumCPU()

		// Create buildenv conf file with default values.
		bytes, err := json.MarshalIndent(b, "", "    ")
		if err != nil {
			return fmt.Errorf("cannot marshal buildenv conf: %w", err)
		}
		if err := os.WriteFile(buildEnvPath, bytes, os.ModePerm); err != nil {
			return err
		}

		return nil
	}

	// Rewrite buildenv file with new platform.
	bytes, err := os.ReadFile(buildEnvPath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bytes, b); err != nil {
		return err
	}

	// Validate cache dirs.
	for index, item := range b.configData.CacheDirs {
		if err := item.Validate(); err != nil {
			return fmt.Errorf("cache dir %d: %w", index, err)
		}
	}

	// Init platform with platform name.
	if err := b.platform.Init(b, b.configData.PlatformName); err != nil {
		return err
	}

	// Init project with project name.
	if err := b.project.Init(b, b.configData.ProjectName); err != nil {
		return err
	}

	return nil
}

func (b buildenv) writeTools(toolchain, environment *strings.Builder) error {
	toolchain.WriteString("\n# Append `path` of tools into $PATH.\n")
	environment.WriteString("\n# Append `path` of tools into $PATH.\n")

	for _, item := range b.platform.Tools {
		toolPath := filepath.Join(Dirs.ToolsDir, item+".json")
		var tool Tool
		if err := tool.Init(toolPath); err != nil {
			return fmt.Errorf("cannot read tool: %s", toolPath)
		}

		if err := tool.Validate(); err != nil {
			return fmt.Errorf("cannot validate tool: %s", toolPath)
		}

		toolchain.WriteString(fmt.Sprintf("set(ENV{PATH} \"%s\")\n", env.Join(tool.cmakepath, "$ENV{PATH}")))
		environment.WriteString(fmt.Sprintf("export PATH=%s\n", env.Join(tool.cmakepath, "$PATH")))
	}
	return nil
}

func (b buildenv) execute(command, workDir string) (string, error) {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("bash", "-c", command)
	}

	var buffer bytes.Buffer
	cmd.Stdout = &buffer
	cmd.Stderr = &buffer

	cmd.Env = os.Environ()
	cmd.Dir = workDir
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

// ----------------------- Implementation of BuildEnvContext ----------------------- //

func (b buildenv) BuildEnvDir() string {
	return filepath.Join(Dirs.WorkspaceDir, "conf")
}

func (b buildenv) Platform() Platform {
	return b.platform
}

func (b buildenv) Project() Project {
	return b.project
}

func (b buildenv) Toolchain() *Toolchain {
	return b.platform.Toolchain
}

func (b buildenv) RootFS() *RootFS {
	return b.platform.RootFS
}

func (b buildenv) SystemName() string {
	if b.Toolchain() == nil {
		return runtime.GOOS
	}

	return b.Toolchain().SystemName
}

func (b buildenv) SystemProcessor() string {
	if b.Toolchain() == nil {
		return runtime.GOARCH
	}
	return b.Toolchain().SystemProcessor
}

func (b buildenv) Host() string {
	if b.Toolchain() == nil {
		if runtime.GOOS == "windows" {
			return "x86_64-w64-mingw32"
		} else if runtime.GOOS == "darwin" {
			return "x86_64-apple-darwin"
		} else {
			return "x86_64-linux-gnu"
		}
	}
	return b.Toolchain().Host
}

func (b buildenv) ToolchainPrefix() string {
	if b.Toolchain() == nil {
		if runtime.GOOS == "windows" {
			return "x86_64-w64-mingw32-"
		} else if runtime.GOOS == "darwin" {
			return "x86_64-apple-darwin-"
		} else if runtime.GOOS == "linux" {
			return "x86_64-linux-gnu-"
		} else {
			panic("unsupported platform: " + runtime.GOOS)
		}
	}
	return b.Toolchain().ToolchainPrefix
}

func (b buildenv) RootFSPath() string {
	if b.RootFS() == nil {
		return ""
	}
	return b.RootFS().fullpath
}

func (b buildenv) BuildType() string {
	return b.buildType
}

func (b buildenv) JobNum() int {
	return b.configData.JobNum
}

func (b buildenv) CacheDirs() []CacheDir {
	return b.configData.CacheDirs
}
