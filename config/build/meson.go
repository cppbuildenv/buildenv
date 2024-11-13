package build

func NewMeson(config BuildConfig) *meson {
	return &meson{BuildConfig: config}
}

type meson struct {
	BuildConfig
}

func (a meson) Configure(buildType string) error {
	return nil
}

func (a meson) Build() error {
	return nil
}

func (a meson) Install() error {
	return nil
}
