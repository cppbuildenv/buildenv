package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func newGenTypedTargets(libInfos CMakeConfig) *genTypedTargets {
	return &genTypedTargets{
		libInfos: libInfos,
	}
}

type genTypedTargets struct {
	libInfos CMakeConfig
}

func (g *genTypedTargets) generate(installedDir string) error {
	// Set namespace to libName if it is empty.
	if g.libInfos.Namespace == "" {
		g.libInfos.Namespace = g.libInfos.LibName
	}

	if g.libInfos.LibName == "" {
		return fmt.Errorf("lib name is empty")
	}

	if g.libInfos.LibType == "" {
		return fmt.Errorf("lib type is empty")
	}

	if g.libInfos.BuildType == "" {
		return fmt.Errorf("build type is empty")
	}

	g.libInfos.LibType = strings.ToLower(g.libInfos.LibType)
	g.libInfos.SystemName = strings.ToLower(g.libInfos.SystemName)

	// Verify importName for windows and soName for linux.
	switch g.libInfos.SystemName {
	case "windows":
		if g.libInfos.SystemName == "windows" && g.libInfos.LibType == "shared" && g.libInfos.LibImpName == "" {
			return fmt.Errorf("import name is empty for windows shared lib")
		}

	case "linux":
		if g.libInfos.SystemName == "linux" && g.libInfos.LibType == "shared" && g.libInfos.LibSoname == "" {
			return fmt.Errorf("so name is empty for linux shared lib")
		}

	default:
		return fmt.Errorf("unsupported system name: %s", g.libInfos.SystemName)
	}

	template := fmt.Sprintf("templates/targets_type_%s_%s.cmake.in", g.libInfos.SystemName, g.libInfos.LibType)
	bytes, err := templates.ReadFile(template)
	if err != nil {
		return err
	}

	// Replace the placeholders with the actual values.
	libNameUpper := strings.ReplaceAll(g.libInfos.LibName, "-", "_")
	libNameUpper = strings.ToUpper(libNameUpper)

	content := string(bytes)
	content = strings.ReplaceAll(content, "@NAMESPACE@", g.libInfos.Namespace)
	content = strings.ReplaceAll(content, "@LIB_NAME@", g.libInfos.LibName)
	content = strings.ReplaceAll(content, "@LIB_NAME_UPPER@", libNameUpper)
	content = strings.ReplaceAll(content, "@LIB_TYPE@", g.libInfos.LibType)
	content = strings.ReplaceAll(content, "@LIB_FILE_NAME@", g.libInfos.LibFilename)
	content = strings.ReplaceAll(content, "@CMAKE_BUILD_TYPE@", strings.ToLower(g.libInfos.BuildType))
	content = strings.ReplaceAll(content, "@CMAKE_BUILD_TYPE_UPPER@", strings.ToUpper(g.libInfos.BuildType))

	switch g.libInfos.SystemName {
	case "windows":
		content = strings.ReplaceAll(content, "@IMPORT_NAME@", g.libInfos.LibImpName)

	case "linux":
		content = strings.ReplaceAll(content, "@SO_NAME@", g.libInfos.LibSoname)

	default:
		return fmt.Errorf("unsupported system name: %s", g.libInfos.SystemName)
	}

	// Make dirs for writing file.
	fileName := fmt.Sprintf("%v-targets-%s.cmake", g.libInfos.LibName, strings.ToLower(g.libInfos.BuildType))
	filePath := filepath.Join(installedDir, "lib", "cmake", g.libInfos.LibName, fileName)
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	// Do write file.
	if err := os.WriteFile(filePath, []byte(content), os.ModePerm); err != nil {
		return err
	}

	return nil
}
