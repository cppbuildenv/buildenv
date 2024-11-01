package config

import "fmt"

type RootFS struct {
	Url     string    `json:"url"`
	Path    string    `json:"path"`
	EnvVars RootFSEnv `json:"env_vars"`
	None    bool      `json:"none"`
}

type RootFSEnv struct {
	SYSROOT                string   `json:"SYSROOT"`
	PKG_CONFIG_SYSROOT_DIR string   `json:"PKG_CONFIG_SYSROOT_DIR"`
	PKG_CONFIG_PATH        []string `json:"PKG_CONFIG_PATH"`
}

func (r RootFS) Verify() error {
	// If none is true, then rootfs is not required.
	if r.None {
		return nil
	}

	if r.Url == "" {
		return fmt.Errorf("rootfs.url is empty")
	}
	if r.Path == "" {
		return fmt.Errorf("rootfs.path is empty")
	}
	if r.EnvVars.SYSROOT == "" {
		return fmt.Errorf("rootfs.env.SYSROOT is empty")
	}
	if r.EnvVars.PKG_CONFIG_SYSROOT_DIR == "" {
		return fmt.Errorf("rootfs.env.PKG_CONFIG_SYSROOT_DIR is empty")
	}
	if len(r.EnvVars.PKG_CONFIG_PATH) == 0 {
		return fmt.Errorf("rootfs.env.PKG_CONFIG_PATH is empty")
	}

	return nil
}
