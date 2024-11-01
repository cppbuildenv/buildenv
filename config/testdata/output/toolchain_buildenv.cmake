set(CMAKE_SYSTEM_NAME "Linux")
set(CMAKE_SYSTEM_PROCESSOR "aarch64")

# Set sysroot for cross-compile.
set(CMAKE_SYSROOT "rootfs/ti-processor-sdk-rtos-j721e-evm-07_03_00_07")
list(APPEND CMAKE_FIND_ROOT_PATH "rootfs/ti-processor-sdk-rtos-j721e-evm-07_03_00_07")
list(APPEND CMAKE_PREFIX_PATH NEVER "rootfs/ti-processor-sdk-rtos-j721e-evm-07_03_00_07")

# Set pkg-config path for cross-compile.
set(ENV{PKG_CONFIG_SYSROOT_DIR} "ti-processor-sdk-rtos-j721e-evm-07_03_00_07")
set(ENV{PKG_CONFIG_PATH} "ti-processor-sdk-rtos-j721e-evm-07_03_00_07/usr/lib/pkgconfig;ti-processor-sdk-rtos-j721e-evm-07_03_00_07/usr/ext/lib/pkgconfig")

# Set toolchain for cross-compile.
set(CMAKE_C_COMPILER "aarch64-none-linux-gnu-gcc")
set(CMAKE_CXX_COMPILER "aarch64-none-linux-gnu-g++")
set(CMAKE_Fortran_COMPILER "aarch64-none-linux-gnu-gfortran")
set(CMAKE_RANLIB "aarch64-none-linux-gnu-ranlib")
set(CMAKE_AR " aarch64-none-linux-gnu-ar")
set(CMAKE_LINKER "aarch64-none-linux-gnu-ld")
set(CMAKE_NM "aarch64-none-linux-gnu-nm")
set(CMAKE_OBJDUMP "aarch64-none-linux-gnu-objdump")
set(CMAKE_STRIP "aarch64-none-linux-gnu-strip")

# Search programs in the host environment.
set(CMAKE_FIND_ROOT_PATH_MODE_PROGRAM NEVER)

# Search libraries and headers in the target environment.
set(CMAKE_FIND_ROOT_PATH_MODE_LIBRARY ONLY)
set(CMAKE_FIND_ROOT_PATH_MODE_INCLUDE ONLY)
set(CMAKE_FIND_ROOT_PATH_MODE_PACKAGE ONLY)
