package config

type VerifyRequest interface {
	Silent() bool
	BuildType() string
	RepairBuildenv() bool
	InstallPorts() bool
}

type verifyRequest struct {
	silent         bool   // Always called from toolchain.cmake
	buildType      string // CMAKE_BUILD_TYPE, default is 'Release'
	repairBuildenv bool   // Called to check and fix build environment.
	installPorts   bool   // Called to install a 3rd party ports.
}

func (v verifyRequest) Silent() bool {
	return v.silent
}

func (v verifyRequest) BuildType() string {
	return v.buildType
}

func (v verifyRequest) RepairBuildenv() bool {
	return v.repairBuildenv
}

func (v *verifyRequest) SetBuildType(buildType string) *verifyRequest {
	v.buildType = buildType
	return v
}

func (v verifyRequest) InstallPorts() bool {
	return v.installPorts
}

func (v *verifyRequest) SetInstallPorts(installPort bool) *verifyRequest {
	v.installPorts = installPort
	return v
}

func NewVerifyRequest(silent, repairBuildenv, installPorts bool) *verifyRequest {
	return &verifyRequest{
		silent:         silent,
		buildType:      "Release",
		repairBuildenv: repairBuildenv,
		installPorts:   installPorts,
	}
}
