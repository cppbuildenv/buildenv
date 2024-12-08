package generator

import (
	"embed"
	"strings"
)

//go:embed templates
var templates embed.FS

// CMakeConfig is the information of the library.
type CMakeConfig struct {
	Namespace string `json:"namespace"` // if empty, use libName instead
	Libtype   string `json:"libtype"`   // it would be STATIC, SHARED or IMPORTED

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

func (gen CMakeConfig) Generate(installedDir string) error {
	gen.Libtype = strings.ToUpper(gen.Libtype)

	var generators []generate

	if len(gen.Components) == 0 {
		generators = []generate{
			&config{gen},
			&targets{gen},
			&configVersion{gen},
			&targetsBuildType{gen},
		}
	} else {
		generators = []generate{
			&config{gen},
			&configVersion{gen},
			&modules{gen},
			&modulesBuildType{gen},
		}
	}

	for _, gen := range generators {
		if err := gen.generate(installedDir); err != nil {
			return err
		}
	}

	return nil
}
