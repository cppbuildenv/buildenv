package config

import (
	"testing"
)

func TestBuildGFlags(t *testing.T) {
	var port Port
	if err := port.Read("testdata/conf/ports/gflags-v2.2.2.json"); err != nil {
		t.Fatal(err)
	}

	// Change for unit tests.
	port.BuildConfig.SourceDir = "testdata/buildtrees/gflags-v2.2.2/src"
	port.BuildConfig.BuildDir = "testdata/buildtrees/gflags-v2.2.2/x86_64-linux-Release"
	port.BuildConfig.InstalledDir = "testdata/installed/x86_64-linux-Release"

	if err := port.Verify(true); err != nil {
		t.Fatal(err)
	}
}
