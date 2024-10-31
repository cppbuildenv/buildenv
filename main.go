package main

import "buildenv/console/cli"

func main() {
	if exit := cli.Listen(); exit {
		return
	}
}
