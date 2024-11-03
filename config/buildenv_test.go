package config

import "testing"

func TestCreateToolchainFile(t *testing.T) {
	var buildenv BuildEnv
	if err := buildenv.Read("testdata/aarch64-linux-test.json"); err != nil {
		t.Fatal(err)
	}

	if err := buildenv.Verify(true); err != nil {
		t.Fatal(err)
	}

	filePath, err := buildenv.CreateToolchainFile("testdata/output")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("toolchain file created: %s", filePath)
}
