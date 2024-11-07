package buildsystem

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

type BuildSystem interface {
	Clone(repo, ref string) error
	Configure(buildType string) error
	Build() error
	Install() error
}

type BuildConfig struct {
	BuildTool string   `json:"build_tool"`
	Arguments []string `json:"arguments"`

	// Internal fields
	SourceDir    string `json:"-"`
	BuildDir     string `json:"-"`
	InstalledDir string `json:"-"`
	JobNum       int    `json:"-"`
}

func (b BuildConfig) Clone(repo, ref string) error {
	var scripts []string

	// Clone repo or sync repo.
	if pathExists(b.SourceDir) {
		scripts = append(scripts, fmt.Sprintf("git -C %s fetch", b.SourceDir))
		scripts = append(scripts, fmt.Sprintf("git -C %s checkout %s", b.SourceDir, ref))
	} else {
		scripts = append(scripts, fmt.Sprintf("git clone --branch %s --single-branch %s %s", ref, repo, b.SourceDir))
	}

	// Execute scripts.
	for _, script := range scripts {
		if err := b.execute(script); err != nil {
			return err
		}
	}

	return nil
}

func (b BuildConfig) execute(command string) error {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("bash", "-c", command)
	}

	var output, errput bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &errput

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	return !os.IsNotExist(err)
}
