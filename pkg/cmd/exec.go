package cmd

import (
	"buildenv/pkg/color"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type executor struct {
	title   string
	command string
	workDir string
	logPath string
}

func NewExecutor(title, command string) *executor {
	return &executor{
		title:   title,
		command: command,
		workDir: "",
		logPath: "",
	}
}

func (e *executor) SetWorkDir(workDir string) *executor {
	e.workDir = workDir
	return e
}

func (e *executor) SetLogPath(logPath string) *executor {
	e.logPath = logPath
	return e
}

func (e *executor) ExecuteOutput() (string, error) {
	var buffer bytes.Buffer
	if err := e.doExecute(&buffer); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func (e executor) Execute() error {
	if err := e.doExecute(nil); err != nil {
		return err
	}

	return nil
}

func (e executor) doExecute(buffer *bytes.Buffer) error {
	if e.title != "" {
		fmt.Print(color.Sprintf(color.Blue, "\n%s: %s\n\n", e.title, e.command))
	}

	// Create command for windows and unix like.
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", e.command)
	} else {
		cmd = exec.Command("bash", "-c", e.command)
	}

	cmd.Dir = e.workDir
	cmd.Env = os.Environ()

	// Create log file if log path specified.
	if e.logPath != "" {
		if err := os.MkdirAll(filepath.Dir(e.logPath), os.ModeDir|os.ModePerm); err != nil {
			return err
		}
		logFile, err := os.Create(e.logPath)
		if err != nil {
			return err
		}
		defer logFile.Close()

		// Write env variables to log file.
		var buffer bytes.Buffer
		for _, envVar := range cmd.Env {
			buffer.WriteString(envVar + "\n")
		}
		io.WriteString(logFile, fmt.Sprintf("Environment:\n%s\n", buffer.String()))

		// Write command summary as header content of file.
		io.WriteString(logFile, fmt.Sprintf("%s: %s\n\n", e.title, e.command))

		cmd.Stdout = io.MultiWriter(os.Stdout, logFile)
		cmd.Stderr = io.MultiWriter(os.Stderr, logFile)
	} else if buffer != nil {
		cmd.Stdout = buffer
		cmd.Stderr = buffer
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
	}

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
