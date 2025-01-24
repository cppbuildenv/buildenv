package buildsystem

func NewBazel(config BuildConfig) *bazel {
	return &bazel{BuildConfig: config}
}

type bazel struct {
	BuildConfig
}

func (b bazel) Configure(buildType string) error {
	return nil
}

func (b bazel) Build() error {
	return nil
}

func (b bazel) Install(withSudo bool) error {
	return nil
}
