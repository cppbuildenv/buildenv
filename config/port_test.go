package config

import (
	"strings"
	"testing"
)

func TestBuildGFlags(t *testing.T) {
	var port Port
	if err := port.Read("testdata/conf/ports/gflags-v2.2.2.json"); err != nil {
		t.Fatal(err)
	}

	// Change for unit tests.
	port.BuildConfig.SrcDir = "testdata/buildtrees/gflags-v2.2.2/src"
	port.BuildConfig.BuildDir = "testdata/buildtrees/gflags-v2.2.2/x86_64-linux-Release"
	port.BuildConfig.InstalledDir = "testdata/installed/x86_64-linux-Release"

	if err := port.Verify(true); err != nil {
		t.Fatal(err)
	}

	// Test clone.
	scripts := port.generateCloneScripts()
	t.Log("clone script: " + strings.Join(scripts, "\n"))
	if err := port.executeScript(scripts); err != nil {
		t.Fatal(err)
	}

	// Test build & install
	scripts = port.generateCMakeBuildScript()
	t.Log("cmake build script: " + strings.Join(scripts, "\n"))
	if err := port.executeScript(scripts); err != nil {
		t.Fatal(err)
	}
}
