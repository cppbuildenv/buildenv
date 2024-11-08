package buildsystem

import (
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

type ColorWriter struct {
	Writer io.Writer
}

func (cw *ColorWriter) Write(p []byte) (n int, err error) {
	coloredOutput := fmt.Sprintf("\033[34m%s\033[0m", string(p))
	_, err = cw.Writer.Write([]byte(coloredOutput))
	if err != nil {
		return 0, err
	}

	return len(p), nil
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

	colorWriter := ColorWriter{Writer: os.Stdout}
	multiWriter := io.MultiWriter(&colorWriter, logFile)

	cmd.Stdout = multiWriter
	cmd.Stderr = multiWriter

	if err := cmd.Run(); err != nil {
		b.printError(fmt.Sprintf("Error execute command: %s", err.Error()))
		return err
	}

	return nil
}

const (
	redFmt     string = "\033[31m%s\033[0m"
	greenFmt   string = "\033[32m%s\033[0m"
	yellowFmt  string = "\033[33m%s\033[0m"
	blueFmt    string = "\033[34m%s\033[0m"
	magentaFmt string = "\033[35m%s\033[0m"
	cyanFmt    string = "\033[36m%s\033[0m"
)

func (b BuildConfig) printSuccess(message string) {
	fmt.Printf(blueFmt, message)
}

func (b BuildConfig) printError(message string) {
	fmt.Printf(redFmt, message)
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	return !os.IsNotExist(err)
}
