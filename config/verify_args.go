package config

type VerifyArgs interface {
	Silent() bool
	CheckAndRepair() bool
	BuildType() string
	PackagePort() string
}

type verifyArgs struct {
	silent         bool   // Always called from toolchain.cmake
	checkAndRepair bool   // Called to check integrity and fix build environment.
	buildType      string // CMAKE_BUILD_TYPE, default is 'Release'
	packagePort    string // If not empty, it means only this port should be verified.
}

func (p verifyArgs) Silent() bool {
	return p.silent
}

func (p verifyArgs) CheckAndRepair() bool {
	return p.checkAndRepair
}

func (p verifyArgs) BuildType() string {
	return p.buildType
}

func (p verifyArgs) PackagePort() string {
	return p.packagePort
}

func (p *verifyArgs) SetVerifyPort(port string) *verifyArgs {
	p.packagePort = port
	return p
}

func NewVerifyArgs(silent, checkAndRepair bool, buildType string) *verifyArgs {
	return &verifyArgs{
		silent:         silent,
		checkAndRepair: checkAndRepair,
		buildType:      buildType,
	}
}
