package buildsystem

func NewNinja(config BuildConfig) *ninja {
	return &ninja{BuildConfig: config}
}

type ninja struct {
	BuildConfig
}

func (n ninja) Configure(buildType string) error {
	// Replace placeholders with real paths and values.
	n.replaceHolders()
	return nil
}

func (n ninja) Build() error {
	return nil
}

func (n ninja) Install() error {
	return nil
}
