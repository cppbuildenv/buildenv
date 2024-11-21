package generator

import "embed"

//go:embed templates
var templates embed.FS

// CMakeConfig is the information of the library.
type CMakeConfig struct {
	Namespace string `json:"namespace"` // if empty, use libName instead
	LibType   string `json:"lib_type"`  // for example: static, shared

	// It's the name of the binary file.
	// in linux, it would be libyaml-cpp.a or libyaml-cpp.so.0.8.0
	// in windows, it would be yaml-cpp.lib or yaml-cpp.dll
	LibFilename string `json:"lib_filename"`

	LibSoname  string `json:"lib_soname"`  // linux, for example: libyaml-cpp.so.0.8
	LibImpName string `json:"lib_impname"` // windows, for example: yaml-cpp.lib

	SystemName string
	LibName    string
	Version    string
	BuildType  string
}

type generate interface {
	generate(installedDir string) error
}

func (l CMakeConfig) Generate(installedDir string) error {
	generators := []generate{
		newGenConfig(l),
		newGenTargets(l),
		newGenConfigVersion(l),
		newGenTypedTargets(l),
	}

	for _, gen := range generators {
		if err := gen.generate(installedDir); err != nil {
			return err
		}
	}
	return nil
}
