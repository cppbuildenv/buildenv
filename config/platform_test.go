package config

import "testing"

func TestCreateToolchainFile(t *testing.T) {
	var buildenv Platform
	if err := buildenv.Read("testdata/conf/platforms/aarch64-linux-test.json"); err != nil {
		t.Fatal(err)
	}

	Dirs.ToolDir = "testdata/conf/tools" // change for test

	args := VerifyArgs{
		Silent:         false,
		CheckAndRepair: false,
		BuildType:      "Release",
	}

	if err := buildenv.Verify(args); err != nil {
		t.Fatal(err)
	}

	filePath, err := buildenv.CreateToolchainFile("testdata/output")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("toolchain file created: %s", filePath)
}
