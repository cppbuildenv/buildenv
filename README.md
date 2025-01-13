# BuildEnv

For the Chinese version of this README, see [README.zh.md](./README.zh.md).

## Introduction

**BuildEnv** is a Go language-based C/C++ package manager that does not require mastering additional program languages. It is designed to simplify package management with JSON only. This package manager works with CMake, with this you can download and setup toolchain, rootfs and tools automacally, then cross-compilation third-party libraries in multiple architectures.

## Background

For a long time, CMake has only provided functions like `find_package` and `find_program`, but it lacks package management capabilities, especially in the following areas:

1. The acquisition and environment configuration of tools required for compilation, such as toolchains, rootfs, CMake, nasm, etc., all of which need to be manually installed and configured in environment variables.
2. The installation directories of third-party libraries and dependency library search directories are not uniformly managed, requiring manual configuration.
3. In terms of cross-compilation support, CMake allows configuration of the cross-compilation environment by specifying the CMAKE_TOOLCHAIN_FILE, but still requires manual configuration.


## Why Not Use Existing Package Management Tools?

While third-party package management tools like `Conan` and `Vcpkg` are widely used in the community, they do not fully meet certain needs:

- **Conan**: Although powerful, Conan depends on the additional Python language, which increases the learning curve. Conan not only supports CMake but also other build systems like `Meson`, `Makefile`, `MSBuild`, `SScon`, `QMake`, `Bazaar`, etc. This makes its API more deeply abstracted, requiring more time to learn new things. As we all known, many c++ developers are still not very familiar with CMake script, this would increase their learning burden.
  
- **Vcpkg**: Easier to use in comparison, but due to networking issues in China, the experience is poor, and it is almost impossible to use properly. Additionally, Vcpkg's default package management only tracks the latest versions of third-party libraries, which complicates managing specific versions of dependencies within projects.

Furthermore, both Conan and Vcpkg do not effectively manage cross-compilation environments. During cross-compilation for multiple platforms, developers often have to manually configure the toolchain, rootfs, and various tools. This process is not only cumbersome but also error-prone.

## Solution：

To solve the above issues, buildenv emerges as a new tool that solves the following core problems:

1. **Management of third-party library installation dir and library search dir during compilation**:  
    - Set CMAKE_PREFIX_PATH, CMAKE_INSTALL_PREFIX globally for cmake projects.
    - Set --prefix globally for Unix Makefiles.
    - Make the pkg-config files work even if current workspace is moved to another folder.

2. **Automatic management of compilation tools**:   
Toolchain, sysroot, CMake, and other tools can be configured in a platform JSON file. You can let them download or specify an absolute path for them.

3. **Generation of CMake config files**:   
For third-party libraries that do not use CMake as a build system, like sqlite3-config.cmake, buildenv can generate them, which making it easy to integrate them into CMake-based projects.

4. **Support specifying third-party library installation and uninstallation**:  
Supports installing and uninstalling libraries along with their sub-dependencies.

5. **Support sharing build cache**:  
Installed files of third-party can be shared with others by configure `cache_dirs` in buildenv's configure file.

For more detailed information, please refer to the [Docs](./docs/01_how_it_works.md).

## Installation Guide

1. Download the Go SDK.
2. Run go build to compile the program successfully.

## Usage Instructions

```
$ ./buildenv

   Welcome to buildenv!                                   
   Please choose an option from the menu below...         
                                                          
  > 1. Init buildenv with conf repo.                      
    2. Create a new platform.                             
    3. Select your current platform.                      
    4. Create a new project.                              
    5. Select your current project.                       
    6. Create a new tool.                                 
    7. Create a new port.                                 
    8. Integrate buildenv, then you can run it everywhere.
    9. About and usage.                                   
                                                          
    ↑/k up • ↓/j down • q quit • ? more 
```

Select the 9 key and press Enter to enter the usage instructions:

```
Welcome to buildenv ().
---------------------------------------
This is a simple pkg-manager for C/C++.

1. How to use it to build cmake project: 
option1: set(CMAKE_TOOLCHAIN_FILE "/mnt/data/work_phil/Golang/buildenv/scripts/toolchain_file.cmake")
option2: cmake .. -DCMAKE_TOOLCHAIN_FILE=/mnt/data/work_phil/Golang/buildenv/scripts/toolchain_file.cmake

2. How to use it to build makefile project: 
source /mnt/data/work_phil/Golang/buildenv/script/environment

[ctrl+c/q -> quit]
```

## How to Contribute

1.  Fork this repository.
2.  Create a new branch Feat_xxx.
3.  Submit your code changes.
4.  Create a Pull Request.