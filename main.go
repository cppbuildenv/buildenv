package main

import (
	"buildenv/cmd"
)

func main() {
	if exit := cmd.Listen(); exit {
		return
	}
}
