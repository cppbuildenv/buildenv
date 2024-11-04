package config

import "testing"

func TestCreateToolchainFile(t *testing.T) {
	var buildenv BuildEnv
	if err := buildenv.Read("testdata/aarch64-linux-test.json"); err != nil {
		t.Fatal(err)
	}

	repoUrl := "http://192.168.100.25:8083/repository/build_resource"
	if err := buildenv.Verify(repoUrl, true); err != nil {
		t.Fatal(err)
	}

	filePath, err := buildenv.CreateToolchainFile("testdata/output")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("toolchain file created: %s", filePath)
}
