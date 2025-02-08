package config

type SetupArgs interface {
	Silent() bool
	BuildType() string
	RepairBuildenv() bool
	InstallPorts() bool
}

type setupArgs struct {
	silent         bool   // Always called from toolchain.cmake
	buildType      string // CMAKE_BUILD_TYPE, default is 'Release'
	repairBuildenv bool   // Called to check and fix build environment.
	installPorts   bool   // Called to install a 3rd party ports.
}

func (s setupArgs) Silent() bool {
	return s.silent
}

func (s setupArgs) BuildType() string {
	return s.buildType
}

func (s setupArgs) RepairBuildenv() bool {
	return s.repairBuildenv
}

func (s *setupArgs) SetBuildType(buildType string) *setupArgs {
	s.buildType = buildType
	return s
}

func (s setupArgs) InstallPorts() bool {
	return s.installPorts
}

func (s *setupArgs) SetInstallPorts(installPort bool) *setupArgs {
	s.installPorts = installPort
	return s
}

func NewSetupArgs(silent, repairBuildenv, installPorts bool) *setupArgs {
	return &setupArgs{
		silent:         silent,
		buildType:      "Release",
		repairBuildenv: repairBuildenv,
		installPorts:   installPorts,
	}
}
