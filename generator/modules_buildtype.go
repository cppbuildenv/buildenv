package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type modulesBuildType struct {
	config GeneratorConfig
}

func (g *modulesBuildType) generate(installedDir string) error {
	if len(g.config.Components) == 0 {
		return fmt.Errorf("components is empty")
	}

	if g.config.SystemName == "" {
		return fmt.Errorf("system name is empty")
	}

	if g.config.BuildType == "" {
		return fmt.Errorf("build type is empty")
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

	modulesTypeBytes, err := templates.ReadFile("templates/modules-buildtype.cmake.in")
	if err != nil {
		return err
	}

	componentTemplatePath := fmt.Sprintf("templates/%s/component-%s.cmake.in",
		strings.ToLower(g.config.SystemName), strings.ToLower(g.config.Libtype))
	componentBytes, err := templates.ReadFile(componentTemplatePath)
	if err != nil {
		return err
	}

	var sections strings.Builder
	for index, component := range g.config.Components {

		var section = string(componentBytes)
		section = strings.ReplaceAll(section, "@NAMESPACE@", g.config.Namespace)
		section = strings.ReplaceAll(section, "@BUILD_TYPE@", g.config.BuildType)
		section = strings.ReplaceAll(section, "@BUILD_TYPE_UPPER@", strings.ToUpper(g.config.BuildType))
		section = strings.ReplaceAll(section, "@COMPONENT@", component.Component)
		section = strings.ReplaceAll(section, "@FILENAME@", component.Filename)
		section = strings.ReplaceAll(section, "@SONAME@", component.Soname)

		if index == 0 {
			sections.WriteString(section + "\n")
		} else if index == len(g.config.Components)-1 {
			sections.WriteString("\n" + section)
		} else {
			sections.WriteString("\n" + section + "\n")
		}
	}

	content := string(modulesTypeBytes)
	content = strings.ReplaceAll(content, "@BUILD_TYPE@", g.config.BuildType)
	content = strings.ReplaceAll(content, "@COMPONENT_SECTIONS@", sections.String())

	// Make dirs for writing file.
	fileName := fmt.Sprintf("%s-modules-%s.cmake", g.config.Libname, strings.ToLower(g.config.BuildType))
	filePath := filepath.Join(installedDir, "lib", "cmake", g.config.Libname, fileName)
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	// Do write file.
	if err := os.WriteFile(filePath, []byte(content), os.ModePerm); err != nil {
		return err
	}

	return nil
}
