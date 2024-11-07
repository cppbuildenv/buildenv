package buildsystem

func NewMake(config BuildConfig) *cmake {
	return &cmake{BuildConfig: config}
}

type make struct {
	BuildConfig
}

func (m make) Configure(buildType string) error {
	return nil
}

func (m make) Build() error {
	return nil
}

func (m make) Install() error {
	return nil
}
