package buildsystem

func NewNinja(config BuildConfig) *ninja {
	return &ninja{BuildConfig: config}
}

type ninja struct {
	BuildConfig
}

func (n ninja) Configure(buildType string) (string, error) {
	return "", nil
}

func (n ninja) Build() (string, error) {
	return "", nil
}

func (n ninja) Install() (string, error) {
	return "", nil
}

func (n ninja) InstalledFiles(installLogFile string) ([]string, error) {
	return nil, nil
}
