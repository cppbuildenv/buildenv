package buildsystem

import (
	"fmt"
	"os"
)

// CrossTools same with `Toolchain` in config/toolchain.go
// redefine to avoid import cycle.
type CrossTools struct {
	FullPath        string
	SystemName      string
	SystemProcessor string
	Host            string
	RootFS          string
	ToolchainPrefix string
	CC              string
	CXX             string
	FC              string
	RANLIB          string
	AR              string
	LD              string
	NM              string
	OBJDUMP         string
	STRIP           string
	Native          bool
}

func (c CrossTools) SetEnvs() {
	if c.Native {
		return
	}

	// Set env vars only for cross compiling.
	rootfs := os.Getenv("SYSROOT")
	os.Setenv("TOOLCHAIN_PREFIX", c.ToolchainPrefix)
	os.Setenv("HOST", c.Host)
	os.Setenv("CC", fmt.Sprintf("%s --sysroot=%s", c.CC, rootfs))
	os.Setenv("CXX", fmt.Sprintf("%s --sysroot=%s", c.CXX, rootfs))

	if c.FC != "" {
		os.Setenv("FC", c.FC)
	}

	if c.RANLIB != "" {
		os.Setenv("RANLIB", c.RANLIB)
	}

	if c.AR != "" {
		os.Setenv("AR", c.AR)
	}

	if c.LD != "" {
		os.Setenv("LD", fmt.Sprintf("%s --sysroot=%s", c.LD, rootfs))
	}

	if c.NM != "" {
		os.Setenv("NM", c.NM)
	}

	if c.OBJDUMP != "" {
		os.Setenv("OBJDUMP", c.OBJDUMP)
	}

	if c.STRIP != "" {
		os.Setenv("STRIP", c.STRIP)
	}
}

func (CrossTools) ClearEnvs() {
	os.Unsetenv("TOOLCHAIN_PREFIX")
	os.Unsetenv("HOST")
	os.Unsetenv("CC")
	os.Unsetenv("CXX")
	os.Unsetenv("FC")
	os.Unsetenv("RANLIB")
	os.Unsetenv("AR")
	os.Unsetenv("LD")
	os.Unsetenv("NM")
	os.Unsetenv("OBJDUMP")
	os.Unsetenv("STRIP")
}
