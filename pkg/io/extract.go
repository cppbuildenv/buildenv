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
	// Open the tar.gz file.
	file, err := os.Open(tarGzFile)
	if err != nil {
		return fmt.Errorf("failed to open tar.gz file: %w", err)
	}
	defer file.Close()

	// Create a gzip reader.
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	// Step 1: Calculate the total size and check the first layer directory
	var (
		totalSize     int64
		firstLayerDir string
	)

	fileName := filepath.Base(tarGzFile)
	PrintInline(fmt.Sprintf("\rCalculating: %s -------- total size of archive file...", fileName))

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		totalSize += header.Size

		// Extract the first layer directory name
		if firstLayerDir == "" {
			parts := strings.SplitN(header.Name, "/", 2)
			if len(parts) > 1 {
				firstLayerDir = parts[0]
			}
		}
	}

	// Reset the file pointer for extraction.
	file.Seek(0, 0)

	// Step 2: Prepare destination directory
	if err := os.RemoveAll(destDir); err != nil {
		return err
	}

	// Extract the tar.gz file.
	gzReader, err = gzip.NewReader(file)
	if err != nil {
		return err
	}
	tarReader = tar.NewReader(gzReader)

	// Step 3: Extract files
	var extractedSize int64
	var lastProgress int

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Determine the file path:
		// If the first layer directory exists and matches the tar.gz file name, remove it from the path.
		var target string
		if firstLayerDir != "" && strings.HasPrefix(header.Name, firstLayerDir+"/") {
			target = filepath.Join(destDir, strings.TrimPrefix(header.Name, firstLayerDir+"/"))
		} else {
			target = filepath.Join(destDir, header.Name)
		}

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
			content := fmt.Sprintf("Extracting:  %s -------- %d%% (%s/%s)",
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
