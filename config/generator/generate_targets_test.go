package generator

import (
	"os"
	"testing"
)

func TestGenTargets(t *testing.T) {
	config := genTargets{
		config: GeneratorConfig{
			Libname:   "yaml-cpp",
			Namespace: "yaml-cpp",
			Libtype:   "SHARED",
		},
	}

	if err := config.generate("temp/yaml-cpp"); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll("temp"); err != nil {
		t.Fatal(err)
	}
}
