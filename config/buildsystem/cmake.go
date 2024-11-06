package buildsystem

import (
	"fmt"
	"os"
	"strings"
)

func NewCMake(config BuildConfig) *cmake {
	return &cmake{BuildConfig: config}
}

type cmake struct {
	BuildConfig
}

func (c cmake) Configure() error {
	// mkdir build.
	if err := os.MkdirAll(c.BuildDir, os.ModeDir|os.ModePerm); err != nil {
		return err
	}

	// Assemble script.
	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_PREFIX_PATH=%s", c.InstalledDir))
	c.Arguments = append(c.Arguments, fmt.Sprintf("-DCMAKE_INSTALL_PREFIX=%s", c.InstalledDir))
	joinedArgs := strings.Join(c.Arguments, " ")
	configure := fmt.Sprintf("cmake -S %s -B %s %s", c.SourceDir, c.BuildDir, joinedArgs)

	// Execute configure.
	if err := c.execute(configure); err != nil {
		return err
	}

	return nil
}

func (c cmake) Build() error {
	// Assemble script.
	build := fmt.Sprintf("cmake --build %s --parallel %d", c.BuildDir, c.JobNum)

	// Execute build.
	if err := c.execute(build); err != nil {
		return err
	}

	return nil
}

func (c cmake) Install() error {
	// Assemble script.
	build := fmt.Sprintf("cmake --install %s", c.BuildDir)

	// Execute install.
	if err := c.execute(build); err != nil {
		return err
	}

	return nil
}
