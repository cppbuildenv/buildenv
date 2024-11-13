package build

func NewNinja(config BuildConfig) *meson {
	return &meson{BuildConfig: config}
}

type ninja struct {
	BuildConfig
}

func (n ninja) Configure(buildType string) error {
	return nil
}

func (n ninja) Build() error {
	return nil
}

func (n ninja) Install() error {
	return nil
}
