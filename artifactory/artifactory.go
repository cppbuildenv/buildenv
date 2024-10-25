package artifactory

type Artifactory interface {
	Toolchain() string
	Sysroot() string
}
