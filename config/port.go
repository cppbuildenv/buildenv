package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type BuildTool int

type Port struct {
	Repo        string      `json:"repo"`
	Ref         string      `json:"ref"`
	Depedencies []string    `json:"dependencies"`
	BuildConfig BuildConfig `json:"build_config"`
}

type BuildConfig struct {
	BuildTool    string   `json:"build_tool"`
	Arguments    []string `json:"arguments"`
	SrcDir       string   `json:"-"`
	BuildDir     string   `json:"-"`
	InstalledDir string   `json:"-"`
	JobNum       int      `json:"-"`
}

func (p *Port) Read(filePath string) error {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, p); err != nil {
		return err
	}

	portName := strings.ReplaceAll(filepath.Base(p.Repo), ".git", "")

	// Set default build dir and installed dir and also can be changed during units tests.
	p.BuildConfig.BuildDir, _ = filepath.Abs(filepath.Join(WorkspaceDir, "buildtrees", portName, p.Ref, "x86_64-linux-Release"))
	p.BuildConfig.SrcDir, _ = filepath.Abs(filepath.Join(WorkspaceDir, "buildtrees", portName, p.Ref, "src"))
	p.BuildConfig.InstalledDir, _ = filepath.Abs(filepath.Join(WorkspaceDir, "installed", "x86_64-linux-Release"))
	p.BuildConfig.JobNum = 8
	return nil
}

func (p *Port) Verify(checkAndRepair bool) error {
	if p.Repo == "" {
		return fmt.Errorf("port.repo is empty")
	}

	if p.Ref == "" {
		return fmt.Errorf("port.ref is empty")
	}

	if p.BuildConfig.BuildTool == "" {
		return fmt.Errorf("port.build_tool is empty")
	}

	if !checkAndRepair {
		return nil
	}

	if err := p.Clone(); err != nil {
		return err
	}

	if err := p.Build(); err != nil {
		return err
	}

	return nil
}

func (p Port) Clone() error {
	scripts := p.generateCloneScripts()
	return p.executeScript(scripts)
}

func (p Port) Build() error {
	scripts := p.generateCMakeBuildScript()
	return p.executeScript(scripts)
}

func (p Port) generateCloneScripts() []string {
	scripts := make([]string, 0)

	// clone repo or sync repo.
	if pathExists(p.BuildConfig.SrcDir) {
		scripts = append(scripts, fmt.Sprintf("git -C %s fetch", p.BuildConfig.SrcDir))
		scripts = append(scripts, fmt.Sprintf("git -C %s checkout %s", p.BuildConfig.SrcDir, p.Ref))
	} else {
		scripts = append(scripts, fmt.Sprintf("git clone --branch %s --single-branch %s %s", p.Ref, p.Repo, p.BuildConfig.SrcDir))
	}

	return scripts
}

func (p Port) generateCMakeBuildScript() []string {
	scripts := make([]string, 0)
	p.BuildConfig.Arguments = append(p.BuildConfig.Arguments, fmt.Sprintf("-DCMAKE_PREFIX_PATH=%s", p.BuildConfig.InstalledDir))
	p.BuildConfig.Arguments = append(p.BuildConfig.Arguments, fmt.Sprintf("-DCMAKE_INSTALL_PREFIX=%s", p.BuildConfig.InstalledDir))
	args := strings.Join(p.BuildConfig.Arguments, " ")

	if p.BuildConfig.BuildTool == "cmake" {
		scripts = append(scripts, fmt.Sprintf("mkdir -p %s", p.BuildConfig.BuildDir))
		scripts = append(scripts, fmt.Sprintf("cmake -S %s -B %s %s", p.BuildConfig.SrcDir, p.BuildConfig.BuildDir, args))
		scripts = append(scripts, fmt.Sprintf("cmake --build %s --target install --parallel %d", p.BuildConfig.BuildDir, p.BuildConfig.JobNum))
	}
	return scripts
}

func (p Port) executeScript(scripts []string) error {
	for _, script := range scripts {
		fmt.Println(script)

		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/c", script)
		} else {
			cmd = exec.Command("bash", "-c", script)
		}

		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}
