package env

import "os"

type Environment struct {
	CPATH                  string
	CFLAGS                 string
	CXXFLAGS               string
	LDFLAGS                string
	PKG_CONFIG_PATH        string
	PKG_CONFIG_SYSROOT_DIR string
}

func (e *Environment) Backup() {
	e.CPATH = os.Getenv("CPATH")
	e.CFLAGS = os.Getenv("CFLAGS")
	e.CXXFLAGS = os.Getenv("CXXFLAGS")
	e.LDFLAGS = os.Getenv("LDFLAGS")
	e.PKG_CONFIG_PATH = os.Getenv("PKG_CONFIG_PATH")
	e.PKG_CONFIG_SYSROOT_DIR = os.Getenv("PKG_CONFIG_SYSROOT_DIR")
}

func (e *Environment) Rollback() {
	// Rollback CPATH.
	if e.CPATH != "" {
		os.Setenv("CPATH", e.CPATH)
	} else {
		os.Unsetenv("CPATH")
	}
	e.CPATH = ""

	// Rollback CFLAGS.
	if e.CFLAGS != "" {
		os.Setenv("CFLAGS", e.CFLAGS)
	} else {
		os.Unsetenv("CFLAGS")
	}
	e.CFLAGS = ""

	// Rollback CXXFLAGS.
	if e.CXXFLAGS != "" {
		os.Setenv("CXXFLAGS", e.CXXFLAGS)
	} else {
		os.Unsetenv("CXXFLAGS")
	}
	e.CXXFLAGS = ""

	// Rollback CPATH.
	if e.LDFLAGS != "" {
		os.Setenv("LDFLAGS", e.LDFLAGS)
	} else {
		os.Unsetenv("LDFLAGS")
	}
	e.LDFLAGS = ""

	// Rollback PKG_CONFIG_PATH.
	if e.PKG_CONFIG_PATH != "" {
		os.Setenv("PKG_CONFIG_PATH", e.PKG_CONFIG_PATH)
	} else {
		os.Unsetenv("PKG_CONFIG_PATH")
	}
	e.PKG_CONFIG_PATH = ""

	// Rollback PKG_CONFIG_SYSROOT_DIR.
	if e.PKG_CONFIG_SYSROOT_DIR != "" {
		os.Setenv("PKG_CONFIG_SYSROOT_DIR", e.PKG_CONFIG_SYSROOT_DIR)
	} else {
		os.Unsetenv("PKG_CONFIG_SYSROOT_DIR")
	}
	e.PKG_CONFIG_SYSROOT_DIR = ""
}
