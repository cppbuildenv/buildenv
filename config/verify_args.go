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

func (v verifyArgs) Silent() bool {
	return v.silent
}

func (v verifyArgs) CheckAndRepair() bool {
	return v.checkAndRepair
}

func (v verifyArgs) BuildType() string {
	return v.buildType
}

func (v *verifyArgs) SetBuildType(buildType string) *verifyArgs {
	v.buildType = buildType
	return v
}

func NewVerifyArgs(silent, checkAndRepair bool) *verifyArgs {
	return &verifyArgs{
		silent:         silent,
		checkAndRepair: checkAndRepair,
		buildType:      "Release",
	}
}
