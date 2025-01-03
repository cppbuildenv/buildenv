package buildsystem

import (
	"bufio"
	"buildenv/pkg/io"
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

func fixupPkgConfig(installedDir string) error {
	pkgConfigDir := filepath.Join(installedDir, "lib", "pkgconfig")

	if !io.PathExists(pkgConfigDir) {
		return nil
	}

	entities, err := os.ReadDir(pkgConfigDir)
	if err != nil {
		return err
	}

	for _, entity := range entities {
		if strings.HasSuffix(entity.Name(), ".pc") {
			pkgPath := filepath.Join(pkgConfigDir, entity.Name())
			if err := doFixupPkgConfig(pkgPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func doFixupPkgConfig(pkgPath string) error {
	pkgFile, err := os.OpenFile(pkgPath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer pkgFile.Close()

	var buffer bytes.Buffer
	scanner := bufio.NewScanner(pkgFile)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "prefix="):
			if line != "prefix=/" {
				buffer.WriteString("prefix=/" + "\n")
			} else {
				buffer.WriteString(line + "\n")
			}

		case strings.HasPrefix(line, "exec_prefix="):
			if line != "exec_prefix=${prefix}" {
				buffer.WriteString("exec_prefix=${prefix}" + "\n")
			} else {
				buffer.WriteString(line + "\n")
			}

		case strings.HasPrefix(line, "libdir="):
			if line != "libdir=${prefix}/lib" {
				buffer.WriteString("libdir=${prefix}/lib" + "\n")
			} else {
				buffer.WriteString(line + "\n")
			}

		case strings.HasPrefix(line, "includedir="):
			if line != "includedir=${prefix}/include" {
				buffer.WriteString("includedir=${prefix}/include" + "\n")
			} else {
				buffer.WriteString(line + "\n")
			}

		case strings.HasPrefix(line, "Libs:"):
			lineOrigin := strings.ReplaceAll(line, "  ", " ")
			line = strings.TrimPrefix(line, "Libs:")
			line = strings.TrimSpace(line)

			parts := strings.Split(line, " ")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "-L") && part != "-L${libdir}" {
					lineOrigin = strings.ReplaceAll(lineOrigin, part, "-L${libdir}")
				}
			}
			buffer.WriteString(lineOrigin + "\n")

		case strings.HasPrefix(line, "Libs.private:"):
			lineOrigin := strings.ReplaceAll(line, "  ", " ")

			line = strings.TrimPrefix(line, "Libs.private:")
			line = strings.TrimSpace(line)

			parts := strings.Split(line, " ")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "-L") && part != "-L${libdir}" {
					lineOrigin = strings.ReplaceAll(lineOrigin, part, "-L${libdir}")
				}
			}
			buffer.WriteString(lineOrigin + "\n")

		case strings.HasPrefix(line, "Cflags:"):
			lineOrigin := strings.ReplaceAll(line, "  ", " ")
			line = strings.TrimPrefix(line, "Cflags:")

			parts := strings.Split(line, " ")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "-I") && part != "-I${includedir}" {
					lineOrigin = strings.ReplaceAll(lineOrigin, part, "-I${includedir}")
				}
			}
			buffer.WriteString(lineOrigin + "\n")

		default:
			buffer.WriteString(line + "\n")
		}
	}

	if buffer.Len() > 0 {
		os.WriteFile(pkgPath, buffer.Bytes(), os.ModePerm)
	}

	return nil
}
