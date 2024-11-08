package buildsystem

import (
	"buildenv/pkg/color"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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
	var commands []string

	// Clone repo or sync repo.
	if pathExists(b.SourceDir) {
		commands = append(commands, fmt.Sprintf("git -C %s fetch", b.SourceDir))
		commands = append(commands, fmt.Sprintf("git -C %s checkout %s", b.SourceDir, ref))
	} else {
		commands = append(commands, fmt.Sprintf("git clone --branch %s --single-branch %s %s", ref, repo, b.SourceDir))
	}

	// Assemble cloneLogPath.
	cloneLogPath := filepath.Join(filepath.Dir(b.BuildDir), filepath.Base(b.BuildDir)+"-clone.log")

	// Execute clone command.
	for _, command := range commands {
		if err := b.execute(command, cloneLogPath); err != nil {
			return err
		}
	}

	return nil
}

func (b BuildConfig) execute(command, logPath string) error {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("bash", "-c", command)
	}

	// Create log file if not exsit.
	if err := os.MkdirAll(filepath.Dir(logPath), os.ModeDir|os.ModePerm); err != nil {
		return err
	}
	logFile, err := os.Create(logPath)
	if err != nil {
		return err
	}
	defer logFile.Close()

	outWriter := color.NewWriter(os.Stdout, color.BlueFmt)
	cmd.Stdout = io.MultiWriter(outWriter, logFile)

	errWriter := color.NewWriter(os.Stdout, color.RedFmt)
	cmd.Stderr = io.MultiWriter(errWriter, logFile)

	if err := cmd.Run(); err != nil {
		color.Println(color.RedFmt, fmt.Sprintf("Error execute command: %s", err.Error()))
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
