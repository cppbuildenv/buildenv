package buildsystem

import (
	"buildenv/pkg/color"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func SyncRepo(sourceDir, repoRef, libName string) error {
	// Change to source dir to execute git command.
	if err := os.Chdir(sourceDir); err != nil {
		return err
	}

	var commands []string
	commands = append(commands, "git reset --hard && git clean -xfd")
	commands = append(commands, fmt.Sprintf("git -C %s fetch origin", sourceDir))
	commands = append(commands, fmt.Sprintf("git -C %s checkout %s", sourceDir, repoRef))
	commands = append(commands, fmt.Sprintf("git -C %s pull origin %s", sourceDir, repoRef))

	// Execute clone command.
	commandLine := strings.Join(commands, " && ")
	title := fmt.Sprintf("[clone %s]", libName)
	if err := execute(title, commandLine, ""); err != nil {
		return err
	}

	return nil
}

func cherryPick(title, sourceDir string, patches []string) error {
	// Change to source dir to execute git command.
	if err := os.Chdir(sourceDir); err != nil {
		return err
	}

	// Execute patch command.
	var commands []string
	commands = append(commands, "git reset --hard && git clean -xfd")
	commands = append(commands, fmt.Sprintf("git -C %s fetch origin", sourceDir))

	for _, patch := range patches {
		commands = append(commands, fmt.Sprintf("git cherry-pick %s", patch))
	}

	commandLine := strings.Join(commands, " && ")
	if err := execute(title, commandLine, ""); err != nil {
		return err
	}

	return nil
}

func rebase(title, sourceDir, repoRef string, rebaseRefs []string) error {
	// Change to source dir to execute git command.
	if err := os.Chdir(sourceDir); err != nil {
		return err
	}

	var commands []string
	commands = append(commands, "git reset --hard && git clean -xfd")
	commands = append(commands, fmt.Sprintf("git -C %s fetch origin", sourceDir))

	for _, ref := range rebaseRefs {
		commands = append(commands, fmt.Sprintf("git checkout %s", ref))
		commands = append(commands, fmt.Sprintf("git rebase %s", repoRef))
	}

	commandLine := strings.Join(commands, " && ")
	if err := execute(title, commandLine, ""); err != nil {
		return err
	}

	return nil
}

func cleanRepo(repoDir string) error {
	if err := os.Chdir(repoDir); err != nil {
		return err
	}

	title := fmt.Sprintf("[clean %s]", filepath.Base(repoDir))
	if err := execute(title, "git reset --hard && git clean -xfd", ""); err != nil {
		return fmt.Errorf("failed to clean source: %v", err)
	}

	return nil
}

func IsRepoModified(repoDir string) (bool, error) {
	cmd := exec.Command("git", "-C", repoDir, "status", "--porcelain")

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("failed to run git command: %v", err)
	}

	status := strings.TrimSpace(out.String())
	return status != "", nil
}

func execute(title, command, logPath string) error {
	fmt.Print(color.Sprintf(color.Blue, "\n%s: %s\n\n", title, command))

	// Create command for windows and linux.
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("bash", "-c", command)
	}
	cmd.Env = os.Environ()

	// Create log file if log path specified.
	if logPath != "" {
		if err := os.MkdirAll(filepath.Dir(logPath), os.ModeDir|os.ModePerm); err != nil {
			return err
		}
		logFile, err := os.Create(logPath)
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
		io.WriteString(logFile, fmt.Sprintf("%s: %s\n\n", title, command))

		cmd.Stdout = io.MultiWriter(os.Stdout, logFile)
		cmd.Stderr = io.MultiWriter(os.Stderr, logFile)
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
	}

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
