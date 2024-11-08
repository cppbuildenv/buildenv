package color

import (
	"fmt"
	"io"
)

const (
	RedFmt     string = "\033[31m%s\033[0m"
	GreenFmt   string = "\033[32m%s\033[0m"
	YellowFmt  string = "\033[33m%s\033[0m"
	BlueFmt    string = "\033[34m%s\033[0m"
	MagentaFmt string = "\033[35m%s\033[0m"
	CyanFmt    string = "\033[36m%s\033[0m"
)

func NewWriter(w io.Writer, colorFmt string) *Writer {
	return &Writer{
		writer:   w,
		colorFmt: colorFmt,
	}
}

type Writer struct {
	writer   io.Writer
	colorFmt string
}

func (w *Writer) Write(p []byte) (n int, err error) {
	coloredOutput := fmt.Sprintf(w.colorFmt, string(p))
	_, err = w.writer.Write([]byte(coloredOutput))
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

func Print(colorFmt, message string) {
	fmt.Printf(colorFmt, message)
}

func Println(colorFmt, message string) {
	fmt.Printf(colorFmt+"\n", message)
}
