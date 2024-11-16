package config

import (
	"testing"
)

func TestBuildGFlags(t *testing.T) {
	portPath := "testdata/conf/ports/gflags-v2.2.2.json"
	platformName := "aarch64-linux-test"
	buildType := "Release"

	var port Port
	if err := port.Init(portPath, platformName, buildType); err != nil {
		t.Fatal(err)
	}

	// Change for unit tests.
	port.BuildConfigs[0].SourceDir = "testdata/buildtrees/gflags-v2.2.2/src"
	port.BuildConfigs[0].BuildDir = "testdata/buildtrees/gflags-v2.2.2/x86_64-linux-Release"
	port.BuildConfigs[0].InstalledDir = "testdata/installed/x86_64-linux-Release"

	args := VerifyArgs{
		Silent:         false,
		CheckAndRepair: false,
		BuildType:      "Release",
	}

	if err := port.Verify(args); err != nil {
		t.Fatal(err)
	}
}
