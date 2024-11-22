package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type genConfigVersion struct {
	config GeneratorConfig
}

func (g *genConfigVersion) generate(installedDir string) error {
	if g.config.Libname == "" {
		return fmt.Errorf("lib name is empty")
	}

	if g.config.Version == "" {
		g.config.Version = "0.0.0"
	}

	bytes, err := templates.ReadFile("templates/config-version.cmake.in")
	if err != nil {
		return err
	}

	// Replace the placeholders with the actual values.
	content := string(bytes)
	content = strings.ReplaceAll(content, "@VERSION@", g.config.Version)

	// Make dirs for writing file.
	filePath := filepath.Join(installedDir, "lib", "cmake", g.config.Libname, g.config.Libname+"-config-version.cmake")
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	// Do write file.
	if err := os.WriteFile(filePath, []byte(content), os.ModePerm); err != nil {
		return err
	}

	return nil
}
