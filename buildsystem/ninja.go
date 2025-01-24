package buildsystem

func NewNinja(config BuildConfig) *ninja {
	return &ninja{
		cmake: *NewCMake(config, "ninja"),
	}
}

type ninja struct {
	cmake
}

func (n ninja) Configure(buildType string) error {
	return n.cmake.Configure(buildType)
}

func (n ninja) Build() error {
	return n.cmake.Build()
}

func (n ninja) Install() error {
	return n.cmake.Install()
}
