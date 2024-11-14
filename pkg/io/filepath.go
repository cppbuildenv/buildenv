package io

import "os"

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	return !os.IsNotExist(err)
}
