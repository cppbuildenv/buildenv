package cli

import (
	"flag"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func newVersionCmd() *versionCmd {
	return &versionCmd{}
}

var (
	AppName      string // for example: `buildenv`
	AppVersion   string // for example: `1.0.4`
	BuildVersion string // for example: `487`
	BuildDate    string // for example: `2020-08-23 12:01:04`
	GoVersion    string // for example: `go version go1.14.1 windows/amd64`
	BuildMode    string // for example: `debug`or `release`, and default is `debug`
)

type versionCmd struct {
	execute bool
}

func (cmd *versionCmd) register() {
	flag.BoolVar(&cmd.execute, "version", false, "print version.")
}

func (cmd *versionCmd) listen() (handled bool) {
	if cmd.execute {
		cmd.PrintVersion()
		return true
	}

	return false
}

// PrintVersion print versions when app launch
func (cmd *versionCmd) PrintVersion() {
	if AppName != "" {
		fmt.Printf("App Name:\t%s\n", AppName)
	}

	if AppVersion != "" {
		fmt.Printf("App Version:\t%s\n", AppVersion)
	}

	if BuildVersion != "" {
		fmt.Printf("Build Version:\t%s\n", BuildVersion)
	}

	if BuildDate != "" {
		fmt.Printf("Build Date:\t%s\n", BuildDate)
	}

	if GoVersion != "" {
		fmt.Printf("Go Version:\t%s\n", GoVersion)
	}

	if BuildMode != "" {
		fmt.Printf("Build Mode:\t%s\n", BuildMode)
	}
}

func init() {
	if len(AppName) == 0 {
		AppName = "buildenv(debug)"
		AppVersion = "NA"
		BuildVersion = "NA"
		BuildDate = "NA"
		GoVersion = "NA"
		BuildMode = "debug"
	}
}

// ReadVersions read all version info into map and return the map
func ReadVersions(filePath string) (map[string]string, error) {
	commands := []string{
		`-version@appName`,
		`-version@appVersion`,
		`-version@goVersion`,
		`-version@buildVersion`,
		`-version@buildMode`,
	}

	versions := make(map[string]string, 0)
	for _, singleCmd := range commands {
		cmd := exec.Command(filePath, singleCmd)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return nil, err
		}

		if err = cmd.Start(); err != nil {
			return nil, err
		}

		bytes, err := io.ReadAll(stdout)
		if err != nil {
			return nil, err
		}

		_ = stdout.Close()
		versions[strings.ReplaceAll(singleCmd, "-version@", "")] = string(bytes)
	}

	return versions, nil
}
