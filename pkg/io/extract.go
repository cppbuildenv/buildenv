package io

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Extract(archiveFile string, destDir string) error {
	switch {
	case strings.HasSuffix(archiveFile, ".tar.gz"):
		return extractTarGz(archiveFile, destDir)

	default:
		return fmt.Errorf("unsupported archive file type: %s", archiveFile)
	}
}

func extractTarGz(tarGzFile string, destDir string) error {
	file, err := os.Open(tarGzFile)
	if err != nil {
		return fmt.Errorf("failed to open tar.gz file: %w", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	// Calculate the total size of the tar.gz file.
	var (
		totalSize     int64
		extractedSize int64
		lastProgress  int
	)

	PrintInline("\rCalculating: total size of archive file...")

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		totalSize += header.Size
	}

	// Reset the file pointer for extraction.
	file.Seek(0, 0)

	if err := os.RemoveAll(destDir); err != nil {
		return err
	}

	// Extract the tar.gz file.
	gzReader, err = gzip.NewReader(file)
	if err != nil {
		return err
	}
	tarReader = tar.NewReader(gzReader)
	fileName := filepath.Base(tarGzFile)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.ModePerm); err != nil {
				return err
			}

		case tar.TypeReg:
			file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			n, err := io.Copy(file, tarReader)
			if err != nil {
				file.Close()
				return err
			}
			file.Close()
			extractedSize += n

		case tar.TypeSymlink:
			if err := os.Symlink(header.Linkname, target); err != nil {
				return err
			}

		default:
			fmt.Printf("Unknown file type: %c\n", header.Typeflag)
		}

		// Update progress.
		progress := int(float64(extractedSize) / float64(totalSize) * 100)
		if progress > lastProgress {
			lastProgress = int(progress)
			content := fmt.Sprintf("Extracting:  %s ---- %d%% (%s/%s)",
				fileName,
				progress,
				formatSize(extractedSize),
				formatSize(totalSize),
			)

			PrintInline(content)
			if progress == 100 {
				fmt.Println()
			}
		}
	}

	return nil
}
