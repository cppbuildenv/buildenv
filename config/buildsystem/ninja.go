package buildsystem

func NewNinja(config BuildConfig) *cmake {
	return &cmake{BuildConfig: config}
}

type ninja struct {
	BuildConfig
}

func (n ninja) Configure() error {
	return nil
}

func (n ninja) Build() error {
	return nil
}

func (n ninja) Install() error {
	return nil
}
