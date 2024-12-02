package main

import (
	"buildenv/menu/cli"
	"flag"
	"os"
)

func main() {
	if exit := cli.Listen(); exit {
		os.Exit(0)
	}

	flag.Usage()
}
