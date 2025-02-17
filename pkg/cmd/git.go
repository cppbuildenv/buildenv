package cmd

import (
	"bufio"
	"buildenv/pkg/fileio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func SyncRepo(sourceDir, repoRef, libName string) error {
	var commands []string
	commands = append(commands, "git reset --hard && git clean -xfd")
	commands = append(commands, fmt.Sprintf("git -C %s fetch origin", sourceDir))
	commands = append(commands, fmt.Sprintf("git -C %s checkout %s", sourceDir, repoRef))
	commands = append(commands, fmt.Sprintf("git -C %s pull origin %s", sourceDir, repoRef))

	// Execute clone command.
	commandLine := strings.Join(commands, " && ")
	title := fmt.Sprintf("[sync %s]", libName)
	executor := NewExecutor(title, commandLine)
	executor.SetWorkDir(sourceDir)
	if err := executor.Execute(); err != nil {
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
	if fileio.PathExists(filepath.Join(repoDir, ".git")) {
		title := fmt.Sprintf("[clean %s]", filepath.Base(repoDir))
		commandLine := "git reset --hard && git clean -xfd"
		executor := NewExecutor(title, commandLine)
		executor.SetWorkDir(repoDir)
		if err := executor.Execute(); err != nil {
			return fmt.Errorf("failed to clean source: %v", err)
		}
	}

	return nil
}

func ApplyPatch(repoDir, patchFile string) error {
	// Check if patched already.
	patchedFlagFile := filepath.Join(repoDir, ".patched")
	if fileio.PathExists(patchedFlagFile) {
		return nil
	}

	file, err := os.Open(patchFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read the first few lines of the file to check for Git patch features.
	var gitBatch bool
	scanner := bufio.NewScanner(file)
	for i := 0; i < 20; i++ {
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()

		// If you find Git patch features such as "From " or "Subject: "
		if strings.HasPrefix(line, "diff --git ") {
			gitBatch = true
			break
		}
	}

	if gitBatch {
		command := fmt.Sprintf("git apply %s", patchFile)
		title := fmt.Sprintf("[patch %s]", filepath.Base(patchFile))
		executor := NewExecutor(title, command)
		executor.SetWorkDir(repoDir)
		if err := executor.Execute(); err != nil {
			return err
		}
	} else {
		// Others, assume it's a regular patch file.
		command := fmt.Sprintf("patch -Np1 -i %s", patchFile)
		title := fmt.Sprintf("[patch %s]", filepath.Base(patchFile))
		executor := NewExecutor(title, command)
		executor.SetWorkDir(repoDir)
		if err := executor.Execute(); err != nil {
			return err
		}
	}

	// Create a flag file to indicated that patch already applied.
	flagFile, err := os.Create(patchedFlagFile)
	if err != nil {
		return fmt.Errorf("cannot create .patched: %w", err)
	}
	defer flagFile.Close()

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
