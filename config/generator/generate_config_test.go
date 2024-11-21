package generator

import (
	"os"
	"testing"
)

func TestGenerateConfig(t *testing.T) {
	config := genConfig{
		libInfos: CMakeConfig{
			Namespace: "yaml-cpp",
			LibName:   "yaml-cpp",
		},
	}

	if err := config.generate("temp/yaml-cpp"); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll("temp"); err != nil {
		t.Fatal(err)
	}
}
