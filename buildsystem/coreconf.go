package buildsystem

func NewCoreConf(config BuildConfig) *coreconf {
	return &coreconf{BuildConfig: config}
}

type coreconf struct {
	BuildConfig
}

func (c coreconf) Configure(buildType string) error {
	return nil
}

func (c coreconf) Build() error {
	return nil
}

func (c coreconf) Install() error {
	return nil
}
