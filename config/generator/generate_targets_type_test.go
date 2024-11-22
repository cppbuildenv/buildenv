package generator

import (
	"os"
	"testing"
)

func TestGenTargetsType(t *testing.T) {
	target := genTargetsBuildType{
		config: GeneratorConfig{
			SystemName: "Linux",
			Namespace:  "yaml-cpp",
			Libname:    "yaml-cpp",
			Libtype:    "SHARED",
			BuildType:  "Release",
			Filename:   "libyaml-cpp.so.0.8.0",
			Soname:     "libyaml-cpp.0.8",
		},
	}

	if err := target.generate("temp/yaml-cpp"); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll("temp"); err != nil {
		t.Fatal(err)
	}
}
