package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type targetsBuildType struct {
	config GeneratorConfig
}

func (g *targetsBuildType) generate(installedDir string) error {
	// Set namespace to libName if it is empty.
	if g.config.Namespace == "" {
		g.config.Namespace = g.config.Libname
	}

	if g.config.Libname == "" {
		return fmt.Errorf("lib name is empty")
	}

	if g.config.Libtype == "" {
		return fmt.Errorf("lib type is empty")
	}

	if g.config.BuildType == "" {
		return fmt.Errorf("build type is empty")
	}

	g.config.Libtype = strings.ToLower(g.config.Libtype)
	g.config.SystemName = strings.ToLower(g.config.SystemName)

	// Verify importName for windows and soName for linux.
	switch g.config.SystemName {
	case "windows":
		if g.config.SystemName == "windows" && g.config.Libtype == "shared" && g.config.Impname == "" {
			return fmt.Errorf("import name is empty for windows shared lib")
		}

	case "linux":
		if g.config.SystemName == "linux" && g.config.Libtype == "shared" && g.config.Soname == "" {
			return fmt.Errorf("so name is empty for linux shared lib")
		}

	default:
		return fmt.Errorf("unsupported system name: %s", g.config.SystemName)
	}

	template := fmt.Sprintf("templates/%s/targets-buildtype-%s.cmake.in", g.config.SystemName, g.config.Libtype)
	bytes, err := templates.ReadFile(template)
	if err != nil {
		return err
	}

	// Replace the placeholders with the actual values.
	libNameUpper := strings.ReplaceAll(g.config.Libname, "-", "_")
	libNameUpper = strings.ToUpper(libNameUpper)

	content := string(bytes)
	content = strings.ReplaceAll(content, "@NAMESPACE@", g.config.Namespace)
	content = strings.ReplaceAll(content, "@LIBNAME@", g.config.Libname)
	content = strings.ReplaceAll(content, "@LIBNAME_UPPER@", libNameUpper)
	content = strings.ReplaceAll(content, "@LIBTYPE@", g.config.Libtype)
	content = strings.ReplaceAll(content, "@FILENAME@", g.config.Filename)
	content = strings.ReplaceAll(content, "@BUILD_TYPE@", strings.ToLower(g.config.BuildType))
	content = strings.ReplaceAll(content, "@BUILD_TYPE_UPPER@", strings.ToUpper(g.config.BuildType))

	switch g.config.SystemName {
	case "windows":
		content = strings.ReplaceAll(content, "@IMPNAME@", g.config.Impname)

	case "linux":
		content = strings.ReplaceAll(content, "@SO_NAME@", g.config.Soname)

	default:
		return fmt.Errorf("unsupported system name: %s", g.config.SystemName)
	}

	// Make dirs for writing file.
	fileName := fmt.Sprintf("%v-targets-%s.cmake", g.config.Libname, strings.ToLower(g.config.BuildType))
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
