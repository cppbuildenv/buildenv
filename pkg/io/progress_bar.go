package io

import "fmt"

type ProgressBar struct {
	fileName     string
	fileSize     int64
	currentSize  int64
	width        int
	lastProgress int
}

func NewProgressBar(fileName string, fileSize int64) *ProgressBar {
	return &ProgressBar{
		fileName: fileName,
		fileSize: fileSize,
		width:    50,
	}
}

func (p *ProgressBar) Write(b []byte) (int, error) {
	n := len(b)
	p.currentSize += int64(n)
	progress := int(float64(p.currentSize*100) / float64(p.fileSize))

	if progress > p.lastProgress {
		p.lastProgress = progress

		fmt.Printf("\rDownloading: %d%% (%s/%s): %s",
			progress,
			p.formatBytes(p.currentSize),
			p.formatBytes(p.fileSize),
			p.fileName)

		if progress == 100 {
			fmt.Println()
		}
	}

	return n, nil
}

func (p ProgressBar) formatBytes(byte int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	var size string

	if byte < KB {
		size = fmt.Sprintf("%.2fB", float64(byte))
	} else if byte < MB {
		size = fmt.Sprintf("%.2fKB", float64(byte)/KB)
	} else if byte < GB {
		size = fmt.Sprintf("%.2fMB", float64(byte)/MB)
	} else {
		size = fmt.Sprintf("%.2fGB", float64(byte)/GB)
	}

	// Remove trailing zeros
	if idx := len(size) - 1; size[idx] == '0' || size[idx-1] == '.' {
		size = size[:idx-1]
	} else if idx := len(size) - 2; size[idx] == '0' {
		size = size[:idx]
	}

	return size
}
