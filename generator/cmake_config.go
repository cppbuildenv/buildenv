package generator

import (
	"buildenv/pkg/io"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed templates
var templates embed.FS

type CMakeConfigs struct {
	Namespace string `json:"namespace"`

	LinuxStatic   CMakeConfig `json:"linux_static"`
	LinuxShared   CMakeConfig `json:"linux_shared"`
	WindowsStatic CMakeConfig `json:"windows_static"`
	WindowsShared CMakeConfig `json:"windows_shared"`
}

func FindMatchedConfig(portDir, configRefer string) (*CMakeConfig, error) {
	configPath := filepath.Join(portDir, "cmake_config.json")
	if !io.PathExists(configPath) {
		return nil, nil
	}
	if strings.TrimSpace(configRefer) == "" {
		return nil, nil
	}

	bytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cmakeConfigs CMakeConfigs
	if err := json.Unmarshal(bytes, &cmakeConfigs); err != nil {
		return nil, err
	}

	var cmakeConfig *CMakeConfig

	switch configRefer {
	case "linux_static":
		cmakeConfig = &cmakeConfigs.LinuxStatic
		cmakeConfig.Libtype = "STATIC"

	case "linux_shared":
		cmakeConfig = &cmakeConfigs.LinuxShared
		cmakeConfig.Libtype = "SHARED"

	case "windows_static":
		cmakeConfig = &cmakeConfigs.WindowsStatic
		cmakeConfig.Libtype = "STATIC"

	case "windows_shared":
		cmakeConfig = &cmakeConfigs.WindowsShared
		cmakeConfig.Libtype = "SHARED"

	default:
		return nil, fmt.Errorf("unknown config refer: %s", configRefer)
	}

	cmakeConfig.Namespace = cmakeConfigs.Namespace
	return cmakeConfig, nil
}

// CMakeConfig is the information of the library.
type CMakeConfig struct {
	// It's the name of the binary file.
	// in linux, it would be libyaml-cpp.a or libyaml-cpp.so.0.8.0
	// in windows, it would be yaml-cpp.lib or yaml-cpp.dll
	Filename string `json:"filename"`

	Soname  string `json:"soname"`  // linux, for example: libyaml-cpp.so.0.8
	Impname string `json:"impname"` // windows, for example: yaml-cpp.lib

	Components []Component `json:"components"`

	// Internal fields.
	Namespace  string // if empty, use libName instead
	SystemName string // for example: Linux, Windows or Darwin
	Libname    string
	Version    string
	BuildType  string
	Libtype    string // it would be STATIC, SHARED or IMPORTED
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
