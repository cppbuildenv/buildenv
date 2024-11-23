package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type genModules struct {
	config GeneratorConfig
}

func (g *genModules) generate(installedDir string) error {
	if len(g.config.Components) == 0 {
		return fmt.Errorf("components is empty")
	}

	if g.config.Libname == "" {
		return fmt.Errorf("lib name is empty")
	}

	if g.config.Version == "" {
		return fmt.Errorf("version is empty")
	}

	if g.config.Libtype == "" {
		return fmt.Errorf("lib type is empty")
	}

	if g.config.Namespace == "" {
		g.config.Namespace = g.config.Libname
	}

	modules, err := templates.ReadFile("templates/modules.cmake.in")
	if err != nil {
		return err
	}

	addLibraryDepedencies, err := templates.ReadFile("templates/modules-addlibrary-depedencies.cmake.in")
	if err != nil {
		return err
	}

	addLibraryIndependent, err := templates.ReadFile("templates/modules-addlibrary-independent.cmake.in")
	if err != nil {
		return err
	}

	var libNames []string
	var addLibrarySections strings.Builder

	for index, component := range g.config.Components {
		libNames = append(libNames, g.config.Libname+"::"+component.Component)

		var section string
		if len(component.Dependencies) > 0 {
			section = string(addLibraryDepedencies)
		} else {
			section = string(addLibraryIndependent)
		}

		var dependencies []string
		for _, dependency := range component.Dependencies {
			dependencies = append(dependencies, g.config.Libname+"::"+dependency)
		}

		section = strings.ReplaceAll(section, "@NAMESPACE@", g.config.Namespace)
		section = strings.ReplaceAll(section, "@LIBNAME@", g.config.Libname)
		section = strings.ReplaceAll(section, "@COMPONENT@", component.Component)
		section = strings.ReplaceAll(section, "@LIBTYPE_UPPER@", strings.ToUpper(g.config.Libtype))
		section = strings.ReplaceAll(section, "@DEPEDENCIES@", strings.Join(dependencies, ";"))

		if index == 0 {
			addLibrarySections.WriteString(section + "\n")
		} else if index == len(g.config.Components)-1 {
			addLibrarySections.WriteString("\n" + section)
		} else {
			addLibrarySections.WriteString("\n" + section + "\n")
		}
	}

	// Replace the placeholders with the actual values.
	libNameUpper := strings.ReplaceAll(g.config.Libname, "-", "_")
	libNameUpper = strings.ToUpper(libNameUpper)

	content := string(modules)
	content = strings.ReplaceAll(content, "@NAMESPACE@", g.config.Namespace)
	content = strings.ReplaceAll(content, "@LIB_NAMES@", strings.Join(libNames, " "))
	content = strings.ReplaceAll(content, "@LIBNAME@", g.config.Libname)
	content = strings.ReplaceAll(content, "@LIBNAME_UPPER@", libNameUpper)
	content = strings.ReplaceAll(content, "@ADD_LIBRARY_SECTIONS@", addLibrarySections.String())

	// Make dirs for writing file.
	filePath := filepath.Join(installedDir, "lib", "cmake", g.config.Libname, g.config.Libname+"-modules.cmake")
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	// Do write file.
	if err := os.WriteFile(filePath, []byte(content), os.ModePerm); err != nil {
		return err
	}

	return nil
}
