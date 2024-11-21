package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func newGenConfig(libInfos CMakeConfig) *genConfig {
	return &genConfig{libInfos}
}

type genConfig struct {
	libInfos CMakeConfig
}

func (g *genConfig) generate(installedDir string) error {
	if g.libInfos.LibName == "" {
		return fmt.Errorf("lib name is empty")
	}

	// Set namespace to libName if it is empty.
	if g.libInfos.Namespace == "" {
		g.libInfos.Namespace = g.libInfos.LibName
	}

	bytes, err := templates.ReadFile("templates/config.cmake.in")
	if err != nil {
		return err
	}

	// Replace the placeholders with the actual values.
	libNameUpper := strings.ReplaceAll(g.libInfos.LibName, "-", "_")
	libNameUpper = strings.ToUpper(libNameUpper)

	content := string(bytes)
	content = strings.ReplaceAll(content, "@LIB_NAME@", g.libInfos.LibName)
	content = strings.ReplaceAll(content, "@LIB_NAME_UPPER@", libNameUpper)
	content = strings.ReplaceAll(content, "@NAMESPACE@", g.libInfos.Namespace)

	// Make dirs for writing file.
	filePath := filepath.Join(installedDir, "lib", "cmake", g.libInfos.LibName, g.libInfos.LibName+"-config.cmake")
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	// Do write file.
	if err := os.WriteFile(filePath, []byte(content), os.ModePerm); err != nil {
		return err
	}

	return nil
}
