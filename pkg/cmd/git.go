package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	if err := NewExecutor(title, commandLine).Execute(); err != nil {
		return err
	}

	return nil
}

func CherryPick(title, sourceDir string, patches []string) error {
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
	if err := NewExecutor(title, commandLine).Execute(); err != nil {
		return err
	}

	return nil
}

func Rebase(title, sourceDir, repoRef string, rebaseRefs []string) error {
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
	if err := NewExecutor(title, commandLine).Execute(); err != nil {
		return err
	}

	return nil
}

func CleanRepo(repoDir string) error {
	if err := os.Chdir(repoDir); err != nil {
		return err
	}

	title := fmt.Sprintf("[clean %s]", filepath.Base(repoDir))
	commandLine := "git reset --hard && git clean -xfd"
	if err := NewExecutor(title, commandLine).Execute(); err != nil {
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
