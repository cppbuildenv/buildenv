package config

type VerifyArgs struct {
	Silent         bool   // Always called from toolchain.cmake
	CheckAndRepair bool   // Called to check integrity and fix build environment.
	BuildType      string // CMAKE_BUILD_TYPE, default is 'Release'
}
