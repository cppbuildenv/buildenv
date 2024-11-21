package generator

import (
	"os"
	"testing"
)

func TestGenerateTargets(t *testing.T) {
	config := genTargets{
		libInfos: CMakeConfig{
			LibName:   "yaml-cpp",
			Namespace: "yaml-cpp",
			LibType:   "SHARED",
		},
	}

	if err := config.generate("temp/yaml-cpp"); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll("temp"); err != nil {
		t.Fatal(err)
	}
}
