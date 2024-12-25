package config

import "testing"

func TestCreateToolchainFile(t *testing.T) {
	platformPath := "testdata/conf/platforms/aarch64-linux-test.json"

	var buildEnv = NewBuildEnv("Release")
	var platform Platform
	if err := platform.Init(buildEnv, platformPath); err != nil {
		t.Fatal(err)
	}

	Dirs.ToolsDir = "testdata/conf/tools" // change for test

	args := NewVerifyArgs(false, false, "Release")
	if err := platform.Verify(args); err != nil {
		t.Fatal(err)
	}

	filePath, err := buildEnv.GenerateToolchainFile("testdata/output")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("toolchain file created: %s", filePath)
}
