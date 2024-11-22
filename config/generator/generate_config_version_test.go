package generator

import (
	"os"
	"testing"
)

func TestGenConfigVersion(t *testing.T) {
	config := genConfigVersion{
		config: GeneratorConfig{
			Libname: "yaml-cpp",
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
