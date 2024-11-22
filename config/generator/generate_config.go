package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type genConfig struct {
	config GeneratorConfig
}

func (g *genConfig) generate(installedDir string) error {
	if g.config.Libname == "" {
		return fmt.Errorf("lib name is empty")
	}

	// Set namespace to libName if it is empty.
	if g.config.Namespace == "" {
		g.config.Namespace = g.config.Libname
	}

	bytes, err := templates.ReadFile("templates/config.cmake.in")
	if err != nil {
		return err
	}

	// Replace the placeholders with the actual values.
	libNameUpper := strings.ReplaceAll(g.config.Libname, "-", "_")
	libNameUpper = strings.ToUpper(libNameUpper)

	content := string(bytes)
	content = strings.ReplaceAll(content, "@LIBNAME@", g.config.Libname)
	content = strings.ReplaceAll(content, "@LIBNAME_UPPER@", libNameUpper)
	content = strings.ReplaceAll(content, "@NAMESPACE@", g.config.Namespace)

	// Make dirs for writing file.
	filePath := filepath.Join(installedDir, "lib", "cmake", g.config.Libname, g.config.Libname+"-config.cmake")
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	// Do write file.
	if err := os.WriteFile(filePath, []byte(content), os.ModePerm); err != nil {
		return err
	}

	return nil
}
