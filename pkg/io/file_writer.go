package io

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type FileWriter struct {
	fileName string
	fileSize int64
	progress func(percent int)
}

func (f FileWriter) Write(reader io.Reader, destDir string) error {
	// Mkdir parent folders if not exist
	if err := os.MkdirAll(destDir, os.ModeDir|os.ModePerm); err != nil {
		return err
	}

	// Create empty file.
	filePath := filepath.Join(destDir, f.fileName)
	dstFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	progressWriter := ProgressWriter{
		Writer:   dstFile,
		fileSize: f.fileSize,
		progress: f.progress,
	}

	// Write data to file.
	if _, err := io.Copy(&progressWriter, reader); err != nil {
		os.Remove(filePath)
		return fmt.Errorf("copy data: %w", err)
	}

	return nil
}

type ProgressWriter struct {
	io.Writer
	fileSize     int64     // Total size of data being written
	lastProgress int       // Last reported progress
	progress     func(int) // Callback function to report progress
	totalWritten int64     // Total bytes written so far
}

// Write method writes data and updates progress.
func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.Writer.Write(p)
	if err != nil {
		return
	}

	pw.totalWritten += int64(n)

	// Calculate progress percentage
	if pw.fileSize > 0 && pw.progress != nil {
		progress := int(float64(pw.totalWritten) / float64(pw.fileSize) * 100.0)
		if progress > pw.lastProgress {
			pw.lastProgress = int(progress)
			pw.progress(progress)
		}
	}

	return
}

func NewFileWriter(fileName string, fileSize int64, progress func(percent int)) FileWriter {
	return FileWriter{
		fileName: fileName,
		fileSize: fileSize,
		progress: progress,
	}
}
