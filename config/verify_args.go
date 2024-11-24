package config

type VerifyArgs interface {
	Silent() bool
	CheckAndRepair() bool
	BuildType() string
}

type verifyArgs struct {
	silent         bool   // Always called from toolchain.cmake
	checkAndRepair bool   // Called to check integrity and fix build environment.
	buildType      string // CMAKE_BUILD_TYPE, default is 'Release'
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

func NewVerifyArgs(silent, checkAndRepair bool, buildType string) *verifyArgs {
	return &verifyArgs{
		silent:         silent,
		checkAndRepair: checkAndRepair,
		buildType:      buildType,
	}
}
