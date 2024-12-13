package buildsystem

func NewMeson(config BuildConfig) *meson {
	return &meson{BuildConfig: config}
}

type meson struct {
	BuildConfig
}

func (a meson) Configure(buildType string) (string, error) {
	return "", nil
}

func (a meson) Build() (string, error) {
	return "", nil
}

func (a meson) Install() (string, error) {
	return "", nil
}

func (m meson) InstalledFiles(installLogFile string) ([]string, error) {
	return nil, nil
}
