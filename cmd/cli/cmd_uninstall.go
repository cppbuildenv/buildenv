package cli

import (
	"bufio"
	"buildenv/config"
	"buildenv/pkg/io"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func newUninstallCmd() *uninstallCmd {
	return &uninstallCmd{}
}

type uninstallCmd struct {
	uninstall string
}

func (u *uninstallCmd) register() {
	flag.StringVar(&u.uninstall, "uninstall", "", "uninstall a 3rd party port.")
}

func (u *uninstallCmd) listen() (handled bool) {
	if strings.TrimSpace(u.uninstall) == "" {
		return false
	}

	// Check port config is exists.
	portPath := filepath.Join(config.Dirs.PortsDir, u.uninstall+".json")
	if !io.PathExists(portPath) {
		fmt.Print(config.UninstallFailed(u.uninstall,
			fmt.Errorf("%s can not be found at %s", u.uninstall+".json", config.Dirs.PortsDir)))
		return true
	}

	args := config.NewVerifyArgs(false, false, buildType.buildType)
	buildenv := config.NewBuildEnv(buildType.buildType)
	if err := buildenv.Verify(args); err != nil {
		fmt.Print(config.UninstallFailed(u.uninstall, err))
		return true
	}

	fileName := fmt.Sprintf("%s-%s.list", buildenv.Platform(), buildenv.BuildType())
	installInfoFile := filepath.Join(config.Dirs.InstalledRootDir, "buildenv", "info", u.uninstall+"-"+fileName)
	if !io.PathExists(installInfoFile) {
		fmt.Print(config.UninstallFailed(u.uninstall,
			fmt.Errorf("%s is not installed", u.uninstall)))
		return true
	}

	file, err := os.OpenFile(installInfoFile, os.O_RDONLY, os.ModePerm)
	if err != nil {
		fmt.Print(config.UninstallFailed(u.uninstall, fmt.Errorf("cannot open install info file: %s", err)))
		return true
	}
	defer file.Close()

	// Read line by line to remove installed file.
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		filePath := filepath.Join(config.Dirs.InstalledRootDir, line)
		if err := os.Remove(filePath); err != nil {
			fmt.Print(config.UninstallFailed(u.uninstall, err))
			return true
		}

		fmt.Printf("remove %s\n", filePath)
	}

	if err := os.Remove(installInfoFile); err != nil {
		fmt.Print(config.UninstallFailed(u.uninstall, err))
		return true
	}

	fmt.Print(config.UninstallSuccess(u.uninstall))

	return true
}
