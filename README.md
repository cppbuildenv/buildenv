# BuildEnv

For the Chinese version of this README, see [README.zh.md](./README.zh.md).

## Introduction

BuildEnv is a Go language-based C/C++ package manager that does not require mastering additional scripting languages. It is designed to simplify package management with JSON format. This package manager is built on CMake, complementing it by solving issues related to compilation, package management, and toolchain resource binding in cross-compilation environments across multiple architectures.

## Background

For a long time, CMake has only provided functions like find_package and find_program, but it lacks package management capabilities, especially in the following areas:

1. The acquisition and environment configuration of tools required for compilation, such as toolchains, rootfs, CMake, nasm, etc., all of which need to be manually installed and configured in environment variables.
2. The installation directories of third-party libraries and dependency library search directories are not uniformly managed, requiring manual configuration.
3. In terms of cross-compilation support, CMake allows configuration of the cross-compilation environment by specifying the CMAKE_TOOLCHAIN_FILE, but still requires manual configuration.


## Why Not Use Existing Package Management Tools?

While third-party package management tools like Conan and Vcpkg are widely used in the community, they do not fully meet certain needs:

- Conan: Although powerful, Conan depends on the additional Python language, which increases the learning curve. Conan not only supports CMake but also other build systems like Meson, Makefile, MSBuild, SScon, QMake, Bazaar, etc. This makes its API more deeply abstracted, requiring more time to learn. For developers who already have limited familiarity with CMake, it adds further learning overhead.
  
- Vcpkg: Easier to use in comparison, but due to networking issues in China, the experience is poor, and it is almost impossible to use properly. Additionally, Vcpkg's default package management only tracks the latest versions of third-party libraries, which complicates managing specific versions of dependencies within projects.

Furthermore, both Conan and Vcpkg do not effectively manage cross-compilation environments. During cross-compilation for multiple platforms, developers often have to manually configure the toolchain, rootfs, and various tools. This process is not only cumbersome but also error-prone.

## Solution：

To address the above issues, buildenv emerges as a new tool that solves the following core problems:

1. Management of third-party library installation directories and library search paths during compilation: Unifies the configuration of find_package paths.
2. Automatic management of compilation tools: Automatically downloads tools like toolchain, sysroot, CMake, and others, and configures their environment variables.
3. Generation of CMake configuration files: For third-party libraries that do not use CMake as a build system, buildenv can automatically generate the corresponding CMake config files, making it easy to integrate them into CMake-based projects.
4. Support for specifying third-party library installation and uninstallation: Automatically compiles and installs dependencies, and supports uninstalling libraries along with their sub-dependencies.
5. Support for shared build cache: Configurable shared directories within a local network for hosting and reading build caches.

For more detailed information, please refer to the Docs.

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
option1: set(CMAKE_TOOLCHAIN_FILE "/mnt/data/work_phil/Golang/buildenv/script/toolchain_file.cmake")
option2: cmake .. -DCMAKE_TOOLCHAIN_FILE=/mnt/data/work_phil/Golang/buildenv/script/toolchain_file.cmake

2. How to use it to build makefile project: 
source /mnt/data/work_phil/Golang/buildenv/script/environment

[ctrl+c/q -> quit]
```

## How to Contribute

1.  Fork this repository.
2.  Create a new branch Feat_xxx.
3.  Submit your code changes.
4.  Create a Pull Request.