package generator

import (
	"os"
	"testing"
)

func TestGenerateConfig(t *testing.T) {
	config := genConfig{
		config: GeneratorConfig{
			Namespace: "yaml-cpp",
			Libname:   "yaml-cpp",
		},
	}

	if err := config.generate("temp/yaml-cpp"); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll("temp"); err != nil {
		t.Fatal(err)
	}
}
