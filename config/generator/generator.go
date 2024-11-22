package generator

import "embed"

//go:embed templates
var templates embed.FS

// GeneratorConfig is the information of the library.
type GeneratorConfig struct {
	Namespace string `json:"namespace"` // if empty, use libName instead
	Libtype   string `json:"libtype"`   // for example: static, shared

	// It's the name of the binary file.
	// in linux, it would be libyaml-cpp.a or libyaml-cpp.so.0.8.0
	// in windows, it would be yaml-cpp.lib or yaml-cpp.dll
	Filename string `json:"filename"`

	Soname  string `json:"soname"`  // linux, for example: libyaml-cpp.so.0.8
	Impname string `json:"impname"` // windows, for example: yaml-cpp.lib

	Components []Component `json:"components"`

	SystemName string
	Libname    string
	Version    string
	BuildType  string
}

type Component struct {
	Component    string   `json:"component"`
	Soname       string   `json:"soname"`
	Impname      string   `json:"impname"`
	Filename     string   `json:"filename"`
	Dependencies []string `json:"dependencies"`
}

type generate interface {
	generate(installedDir string) error
}

func (gen GeneratorConfig) Generate(installedDir string) error {
	var generators []generate

	if len(gen.Components) == 0 {
		generators = []generate{
			&genConfig{gen},
			&genTargets{gen},
			&genConfigVersion{gen},
			&genTargetsBuildType{gen},
		}
	} else {
		generators = []generate{
			&genConfig{gen},
			&genConfigVersion{gen},
			&genModules{gen},
			&genModulesBuildType{gen},
		}
	}

	for _, gen := range generators {
		if err := gen.generate(installedDir); err != nil {
			return err
		}
	}

	return nil
}
