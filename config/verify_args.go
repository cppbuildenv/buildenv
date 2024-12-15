package config

type VerifyArgs interface {
	Silent() bool
	CheckAndRepair() bool
	BuildType() string
	PortToInstall() string
}

type verifyArgs struct {
	silent          bool   // Always called from toolchain.cmake
	checkAndRepair  bool   // Called to check integrity and fix build environment.
	buildType       string // CMAKE_BUILD_TYPE, default is 'Release'
	portToInstall   string // If not empty, it means only this port should be installed.
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

func (v verifyArgs) PortToInstall() string {
	return v.portToInstall
}

func (v *verifyArgs) InstallPort(port string) *verifyArgs {
	v.portToInstall = port
	return v
}

func NewVerifyArgs(silent, checkAndRepair bool, buildType string) *verifyArgs {
	return &verifyArgs{
		silent:         silent,
		checkAndRepair: checkAndRepair,
		buildType:      buildType,
	}
}
