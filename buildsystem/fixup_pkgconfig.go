package buildsystem

import (
	"bufio"
	"buildenv/pkg/fileio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

func fixupPkgConfig(packageDir, prefix string) error {
	pkgConfigShareDir := filepath.Join(packageDir, "share", "pkgconfig")
	pkgConfigLibDir := filepath.Join(packageDir, "lib", "pkgconfig")
	pkgConfigLib64Dir := filepath.Join(packageDir, "lib64", "pkgconfig")

	// Move pkg-config files from shared dir to lib dir.
	if fileio.PathExists(pkgConfigShareDir) {
		matched, err := filepath.Glob(filepath.Join(pkgConfigShareDir, "*.pc"))
		if err != nil {
			return err
		}

		if len(matched) > 0 {
			if err := os.MkdirAll(pkgConfigLibDir, os.ModeDir|os.ModePerm); err != nil {
				return err
			}
		}

		// Move pc files from share to lib.
		for _, pkgPath := range matched {
			fileName := filepath.Base(pkgPath)
			if err := os.Rename(pkgPath, filepath.Join(pkgConfigLibDir, fileName)); err != nil {
				return err
			}
		}

		// Remove empty shared dir.
		if err := os.RemoveAll(pkgConfigShareDir); err != nil {
			return err
		}
	}

	// Fixup pkg-config files in `lib` if exists.
	if fileio.PathExists(pkgConfigLibDir) {
		entities, err := os.ReadDir(pkgConfigLibDir)
		if err != nil {
			return err
		}
		for _, entity := range entities {
			if strings.HasSuffix(entity.Name(), ".pc") {
				pkgPath := filepath.Join(pkgConfigLibDir, entity.Name())
				if err := doFixupPkgConfig(pkgPath, prefix); err != nil {
					return err
				}
			}
		}
	}

	// Fixup pkg-config files in `lib64` if exists,
	// but we always try to install libraries into `lib` instead of `lib64`.
	if fileio.PathExists(pkgConfigLib64Dir) {
		entities, err := os.ReadDir(pkgConfigLib64Dir)
		if err != nil {
			return err
		}
		for _, entity := range entities {
			if strings.HasSuffix(entity.Name(), ".pc") {
				pkgPath := filepath.Join(pkgConfigLib64Dir, entity.Name())
				if err := doFixupPkgConfig(pkgPath, prefix); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func doFixupPkgConfig(pkgPath, prefix string) error {
	pkgFile, err := os.OpenFile(pkgPath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer pkgFile.Close()

	var buffer bytes.Buffer
	scanner := bufio.NewScanner(pkgFile)
	for scanner.Scan() {
		line := scanner.Text()

		// Remove space before `=`
		line = strings.ReplaceAll(line, "prefix =", "prefix=")
		line = strings.ReplaceAll(line, "exec_prefix =", "exec_prefix=")
		line = strings.ReplaceAll(line, "libdir =", "libdir=")
		line = strings.ReplaceAll(line, "sharedlibdir =", "sharedlibdir=")
		line = strings.ReplaceAll(line, "includedir =", "includedir=")

		switch {
		case strings.HasPrefix(line, "prefix="):
			if line != "prefix=" {
				buffer.WriteString("prefix=" + prefix + "\n")
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

		case strings.HasPrefix(line, "sharedlibdir="):
			if line != "sharedlibdir=${prefix}/lib" {
				buffer.WriteString("sharedlibdir=${prefix}/lib" + "\n")
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

		default:
			buffer.WriteString(line + "\n")
		}
	}

	if buffer.Len() > 0 {
		os.WriteFile(pkgPath, buffer.Bytes(), os.ModePerm)
	}

	return nil
}
