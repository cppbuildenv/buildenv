package config

import "testing"

func TestGenerateToolchainFile(t *testing.T) {
	var buildenv BuildEnv
	buildenv.Read("testdata/aarch64-linux-jetson-nano.json")

}
