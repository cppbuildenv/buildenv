package pkg

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type BuildEnv struct {
	Host     string  `json:"host"`
	Platform string  `json:"platform"`
	BuildEnv Profile `json:"buildenv"`
}

type Profile struct {
	Arch      string `json:"arch"`
	OS        string `json:"os"`
	Sysroot   string `json:"sysroot"`
	Toolchain string `json:"toolchain"`
}

func (b BuildEnv) Verify() bool {
	if b.Host == "" {
		log.Println("Host is empty")
		return false
	}

	if b.Platform == "" {
		log.Println("Platform is empty")
		return false
	}

	if !b.BuildEnv.Verify() {
		log.Println("BuildEnv is empty")
		return false
	}

	return true
}

func (b BuildEnv) Write(filename string) error {
	bytes, err := json.MarshalIndent(b, "", "    ")
	if err != nil {
		return err
	}

	if !strings.HasSuffix(filename, ".json") {
		filename = filename + ".json"
	}
	return os.WriteFile(filepath.Join("buildenv", filename), bytes, 0644)
}

func (p Profile) Verify() bool {
	if p.Arch == "" {
		log.Println("Arch is empty")
		return false
	}

	if p.OS == "" {
		log.Println("OS is empty")
		return false
	}

	if p.Sysroot == "" {
		log.Println("Sysroot is empty")
		return false
	}

	if p.Toolchain == "" {
		log.Println("Toolchain is empty")
		return false
	}

	return true
}
