package config

import (
	"buildenv/pkg/color"
	"fmt"
)

func SprintSuccess(format string, args ...interface{}) string {
	return color.Sprintf(color.Magenta, "\n[✔] ======== %s ========\n\n", fmt.Sprintf(format, args...))
}

func SprintError(err error, format string, args ...interface{}) string {
	return color.Sprintf(color.Red, "\n[✘] %s\n[☛] %s.\n\n", err, fmt.Sprintf(format, args...))
}

func PrintSuccess(format string, args ...interface{}) {
	color.Printf(color.Magenta, "\n[✔] ======== %s ========\n\n", fmt.Sprintf(format, args...))
}

func PrintError(err error, format string, args ...interface{}) {
	color.Printf(color.Red, "\n[✘] %s\n[☛] %s.\n\n", fmt.Sprintf(format, args...), err)
}
