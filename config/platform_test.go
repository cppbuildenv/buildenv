package config

import "testing"

func TestCreateToolchainFile(t *testing.T) {
	platformPath := "testdata/conf/platforms/aarch64-linux-test.json"

	var buildEnv = NewBuildEnv()
	var platform Platform
	if err := platform.Init(buildEnv, platformPath); err != nil {
		t.Fatal(err)
	}

	Dirs.ToolsDir = "testdata/conf/tools" // change for test

	request := NewVerifyRequest(false, false, false)
	if err := platform.Verify(request); err != nil {
		t.Fatal(err)
	}

	filePath, err := buildEnv.GenerateToolchainFile("testdata/output")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("toolchain file created: %s", filePath)
}
