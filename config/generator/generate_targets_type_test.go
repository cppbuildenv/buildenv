package generator

import (
	"os"
	"testing"
)

func TestGenerateTypedTargets(t *testing.T) {
	target := genTypedTargets{
		libInfos: CMakeConfig{
			SystemName:  "linux",
			Namespace:   "yaml-cpp",
			LibName:     "yaml-cpp",
			LibType:     "shared",
			BuildType:   "release",
			LibFilename: "libyaml-cpp.so.0.8.0",
			LibSoname:   "libyaml-cpp.0.8",
		},
	}

	if err := target.generate("temp/yaml-cpp"); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll("temp"); err != nil {
		t.Fatal(err)
	}
}
