package config

import (
	"testing"
)

func TestBuildGFlags(t *testing.T) {
	buildenv := NewBuildEnv()

	var port Port
	if err := port.Init(buildenv, "gflags@v2.2.2"); err != nil {
		t.Fatal(err)
	}

	if err := port.Validate(); err != nil {
		t.Fatal(err)
	}
}
