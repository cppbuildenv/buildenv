package config

import "fmt"

type Toolchain struct {
	Url           string          `json:"url"`
	Path          string          `json:"path"`
	EnvVars       ToolchainEnvVar `json:"env_vars"`
	ToolChainVars ToolChainVars   `json:"toolchain_vars"`
}

type ToolchainEnvVar struct {
	CC      string `json:"CC"`
	CXX     string `json:"CXX"`
	FC      string `json:"FC"`
	RANLIB  string `json:"RANLIB"`
	AR      string `json:"AR"`
	LD      string `json:"LD"`
	NM      string `json:"NM"`
	OBJDUMP string `json:"OBJDUMP"`
	STRIP   string `json:"STRIP"`
}

type ToolChainVars struct {
	CMAKE_SYSTEM_NAME      string `json:"CMAKE_SYSTEM_NAME"`
	CMAKE_SYSTEM_PROCESSOR string `json:"CMAKE_SYSTEM_PROCESSOR"`
}

func (t Toolchain) Verify() error {
	if t.Url == "" {
		return fmt.Errorf("toolchain.url is empty")
	}

	if t.Path == "" {
		return fmt.Errorf("toolchain.path is empty")
	}

	if t.EnvVars.CC == "" {
		return fmt.Errorf("toolchain.env.CC is empty")
	}

	if t.EnvVars.CXX == "" {
		return fmt.Errorf("toolchain.env.CXX is empty")
	}

	if t.ToolChainVars.CMAKE_SYSTEM_NAME == "" {
		return fmt.Errorf("toolchain.toolchain_vars.CMAKE_SYSTEM_NAME is empty")
	}

	if t.ToolChainVars.CMAKE_SYSTEM_PROCESSOR == "" {
		return fmt.Errorf("toolchain.toolchain_vars.CMAKE_SYSTEM_PROCESSOR is empty")
	}

	return nil
}
