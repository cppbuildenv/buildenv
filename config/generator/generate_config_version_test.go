package generator

import (
	"os"
	"testing"
)

func TestGenerateConfigVersion(t *testing.T) {
	config := genConfigVersion{
		libInfos: CMakeConfig{
			LibName: "yaml-cpp",
			Version: "0.8.0",
		},
	}

	if err := config.generate("temp/yaml-cpp"); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll("temp"); err != nil {
		t.Fatal(err)
	}
}
