package config

import (
	"testing"
)

func TestBuildGFlags(t *testing.T) {
	portPath := "testdata/conf/ports/gflags-v2.2.2.json"
	buildenv := NewBuildEnv()

	var port Port
	if err := port.Init(buildenv, portPath); err != nil {
		t.Fatal(err)
	}

	if err := port.Verify(); err != nil {
		t.Fatal(err)
	}
}
