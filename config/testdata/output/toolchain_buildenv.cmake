set(CMAKE_SYSTEM_NAME "Linux")
set(CMAKE_SYSTEM_PROCESSOR "aarch64")

# Set sysroot for cross-compile.
set(CMAKE_SYSROOT "/mnt/data/work_phil/Golang/buildenv/config/workspace/rootfs/rootfs/ti-processor-sdk-rtos-j721e-evm-07_03_00_07")
list(APPEND CMAKE_FIND_ROOT_PATH "${CMAKE_SYSROOT}")
list(APPEND CMAKE_PREFIX_PATH NEVER "${CMAKE_SYSROOT}")

# Set pkg-config path for cross-compile.
set(ENV{PKG_CONFIG_SYSROOT_DIR} "${CMAKE_SYSROOT}")
set(ENV{PKG_CONFIG_PATH} "/mnt/data/work_phil/Golang/buildenv/config/workspace/rootfs/ti-processor-sdk-rtos-j721e-evm-07_03_00_07/usr/lib/pkgconfig;/mnt/data/work_phil/Golang/buildenv/config/workspace/rootfs/ti-processor-sdk-rtos-j721e-evm-07_03_00_07/usr/ext/lib/pkgconfig")

# Set toolchain for cross-compile.
set(_TOOLCHAIN_BIN_PATH 	"workspace/toolchain/aarch64-linux/gcc-arm-9.2-2019.12-x86_64-aarch64-none-linux-gnu/bin")
set(CMAKE_C_COMPILER 		"${_TOOLCHAIN_BIN_PATH}/aarch64-none-linux-gnu-gcc")
set(CMAKE_CXX_COMPILER		"${_TOOLCHAIN_BIN_PATH}/aarch64-none-linux-gnu-g++")
set(CMAKE_Fortran_COMPILER	"${_TOOLCHAIN_BIN_PATH}/aarch64-none-linux-gnu-gfortran")
set(CMAKE_RANLIB 			"${_TOOLCHAIN_BIN_PATH}/aarch64-none-linux-gnu-ranlib")
set(CMAKE_AR 				"${_TOOLCHAIN_BIN_PATH}/ aarch64-none-linux-gnu-ar")
set(CMAKE_LINKER 			"${_TOOLCHAIN_BIN_PATH}/aarch64-none-linux-gnu-ld")
set(CMAKE_NM 				"${_TOOLCHAIN_BIN_PATH}/aarch64-none-linux-gnu-nm")
set(CMAKE_OBJDUMP 			"${_TOOLCHAIN_BIN_PATH}/aarch64-none-linux-gnu-objdump")
set(CMAKE_STRIP 			"${_TOOLCHAIN_BIN_PATH}/aarch64-none-linux-gnu-strip")

# Search programs in the host environment.
set(CMAKE_FIND_ROOT_PATH_MODE_PROGRAM NEVER)

# Search libraries and headers in the target environment.
set(CMAKE_FIND_ROOT_PATH_MODE_LIBRARY ONLY)
set(CMAKE_FIND_ROOT_PATH_MODE_INCLUDE ONLY)
set(CMAKE_FIND_ROOT_PATH_MODE_PACKAGE ONLY)
