package io

import (
	"fmt"
	"os"
	"testing"
)

func TestExtract7z(t *testing.T) {
	defer os.RemoveAll("temp")

	if err := Extract("testdata/extract.tar.gz", "temp"); err != nil {
		t.Fatal(err)
	}

	if err := verifyExtracted(); err != nil {
		t.Fatal(err)
	}
}

func TestExtractTarGz(t *testing.T) {
	defer os.RemoveAll("temp")

	if err := Extract("testdata/extract.tar.gz", "temp"); err != nil {
		t.Fatal(err)
	}

	if err := verifyExtracted(); err != nil {
		t.Fatal(err)
	}
}

func TestExtractTarXz(t *testing.T) {
	defer os.RemoveAll("temp")

	if err := Extract("testdata/extract.tar.xz", "temp"); err != nil {
		t.Fatal(err)
	}

	if err := verifyExtracted(); err != nil {
		t.Fatal(err)
	}
}

func TestExtractZip(t *testing.T) {
	defer os.RemoveAll("temp")

	if err := Extract("testdata/extract.zip", "temp"); err != nil {
		t.Fatal(err)
	}

	if err := verifyExtracted(); err != nil {
		t.Fatal(err)
	}
}

func verifyExtracted() error {
	files := []string{
		"temp/extract/111.txt",
		"temp/extract/222/222.txt",
		"temp/extract/333/333.txt",
	}

	for _, file := range files {
		if !PathExists(file) {
			return fmt.Errorf("file: %v not exists", file)
		}
	}

	info, err := os.Lstat("temp/extract/333/333.txt")
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("temp/extract/333/333.txt should be symbolic link")
	}

	return nil
}
