package buildsystem

func NewMeson(config BuildConfig) *cmake {
	return &cmake{BuildConfig: config}
}

type meson struct {
	BuildConfig
}

func (a meson) Configure() error {
	return nil
}

func (a meson) Build() error {
	return nil
}

func (a meson) Install() error {
	return nil
}
