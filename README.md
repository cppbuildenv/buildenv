# BuildEnv

## 介绍 - Introduction

buildenv是一个用 **Go语言** 实现的 **C/C++ 包管理器**，不需要掌握额外的脚本语言，只需了解 **JSON** 格式即可轻松管理包。该包管理器基于 **CMake**，作为 **CMake** 的补充，主要解决 **CMake 在多架构平台交叉编译环境下的编译、包管理以及所属工具资源绑定问题**。

-----

**buildenv** is a Go language-based C/C++ package manager that does not require mastering additional program languages. It is designed to simplify package management with JSON only. This package manager works with CMake, with this you can download and setup toolchain, rootfs and tools automacally, then cross-compilation third-party libraries in multiple architectures.


## 背景问题 - Background.

CMake长期以来仅提供了 `find_package`、`find_program`、`find_library`、`find_path` 等功能，但缺乏对包的管理能力，特别是在以下几个方面：

-----

For a long time, CMake has only provided functions like  `find_package`、`find_program`、`find_library`、`find_path`, but it lacks package management capabilities, especially in the following areas:

1. 编译所需要的工具获取和环境配置，如toolchain, rootfs, cmake，nasm等需要手动安装和配置环境变量；

    ----

    The acquisition and environment configuration of tools required for compilation, such as toolchains, rootfs, CMake, nasm, etc., all of which need to be manually installed and configured in environment variables.

2. 三方库编译后安装目录和依赖库寻找目录缺乏统一管理，需要手动配置；

    ----

    The installation directories of third-party libraries and dependency library search directories are not uniformly managed, requiring manual configuration.

3. 交叉编译支持方面，CMake允许通过指定CMAKE_TOOLCHAIN_FILE来配置交叉编译环境，但仍需手动配置。

    ----

    In terms of cross-compilation support, CMake allows configuration of the cross-compilation environment by specifying the CMAKE_TOOLCHAIN_FILE, but still requires manual configuration.

## 为什么不使用现有的包管理工具 - Why Not Use Existing Package Management Tools

尽管 Conan和Vcpkg等第三方包管理工具在社区中已经得到了广泛使用，但它们并不能完全满足某些需求：

----

While third-party package management tools like `Conan` and `Vcpkg` are widely used in the community, they do not fully meet certain needs:

1. **Conan**：虽然功能强大，但依赖于额外的 **Python** 语言和python包，且上手成本较高。因为 Conan 不仅支持 **CMake**，还支持 **Meson**、**Makefile**、**MSBuild**、**SScon**、**QMake**、**Bazaar** 等构建系统，这导致其 API 封装较深，需要更多时间学习和上手，对于本来**CMake**掌握就一般的同学无疑又增加了额外新的学习成本。

    ----

    **Conan**: Although powerful, Conan depends on the additional Python language, which increases the learning curve. Conan not only supports CMake but also other build systems like `Meson`, `Makefile`, `MSBuild`, `SScon`, `QMake`, `Bazaar`, etc. This makes its API more deeply abstracted, requiring more time to learn new things. As we all known, many c++ developers are still not very familiar with CMake script, this would increase their learning burden.
  
2. **Vcpkg**：相对容易上手，但由于 **国内网络环境问题**使用体验较差，几乎无法正常使用，而且**Vcpkg**对于三方库的版本管理过于简单，对于多版本依赖管理不灵活。

    ----

    **Vcpkg**: Easier to use in comparison, but due to networking issues in China, the experience is poor, and it is almost impossible to use properly. Additionally, Vcpkg's default package management is too simplistic, and it is not flexible for managing multiple versions of dependencies.

另外，**Conan** 和 **Vcpkg** 都未能有效管理 **交叉编译环境**，在多个平台的交叉编译时，开发者通常需要手动配置 toolchain 和 rootfs, 以及各种tool的配置，这样不仅繁琐，而且容易出错。

----

Furthermore, both Conan and Vcpkg do not effectively manage cross-compilation environments. During cross-compilation for multiple platforms, developers often have to manually configure the toolchain, rootfs, and various tools. This process is not only cumbersome but also error-prone.

## 解决方案：

为了解决上述问题，**buildenv** 作为一个新的工具应运而生，主要解决以下几个核心问题：

----

To solve the above issues, buildenv emerges as a new tool that solves the following core problems:

1. **支持管理三方库的安装目录以及编译期间依赖库的寻找目录** -  **Management of third-party library installation dir and library search dir during compilation**：
    - 给CMake项目全局设置 `CMAKE_PREFIX_PATH`, `CMAKE_INSTALL_PREFIX`；
    - 给Unix Makefiles项目全局设置 `--prefix`；
    - 让Unix Makefiles项目在编译期间能通过pc文件找到子依赖，即便当前workspace目录迁移了；

    ----

    - Set `CMAKE_PREFIX_PATH`, `CMAKE_INSTALL_PREFIX` globally for cmake projects.
    - Set `--prefix` globally for Unix Makefiles.
    - Make the pkg-config files work even if current workspace is moved to another folder.

2. **支持自动管理编译工具** - **Automatic management of compilation tools**：  
通过配置实现自动下载 `toolchain`、`sysroot`、`CMake`、`ninja`、`nasm` 等以及配置其环境变量；

    ----
   
    Toolchain, sysroot, CMake, and other tools can be configured in a platform JSON file. You can let them download or specify an absolute path for them.

3. **支持生成CMake配置文件** - **Generation of CMake config files**  
对于非CMake作为构建工具的三方库，可以自动生成对应的cmake config文件，方便在CMake项目中使用；

    ----

    For third-party libraries that do not use CMake as a build system, like sqlite3-config.cmake, buildenv can generate them, which making it easy to integrate them into CMake-based projects.

4. **支持指定三方库的install和uninstall** - **Support specifying third-party library installation and uninstallation**:  
自动编译和安装子依赖，支持卸载库同时卸载子依赖；

    ----

    Supports installing and uninstalling libraries along with their sub-dependencies.

5. **支持内部版本冲突检查** - **Support detecting version conflict of same library with different versions in workspace**：
内部版本冲突检查，即检查当前workspace下的三方库是否存在多个版本，若存在多个版本，会提示用户选择一个版本；

    ----
  
    Show error message to warning user if there are multiple versions of the same library in the workspace.

6. **支持编译缓存共享**:  
通过配置`cache_dirs`，可进行局域网内网盘来托管和读取`install文件缓存`；

    ----

    **Support sharing build cache**:  
Installed files of third-party can be shared with others by configure `cache_dirs` in buildenv's configure file.

## 如何编编译 - Build Guide.

1. 下载`golang sdk`；

    ----

    Download the Go SDK.

2. `go build`，即可编译成功；

    ----

    Run go build to compile the program successfully.

3. 或者执行内置的脚本`build.sh`即可编译成功。

    ----

    You can also build it by execute `./build.sh`.

## 使用说明 - Usage Instructions.

buildenv 提供两种交互使用方式：cli和gui，前者便于CI/CD里使用，后者便于本地开发使用，除了cli模式会额外提供`install`和`uninstall`相关命令之外两者的使用方式基本一致, gui模式如下：

---
buildenv providers two kinds of usage: cli and gui, cli mode will provide `install` and `uninstall` commands, and the usage of them is almost the same, except that cli mode will provide `install` and `uninstall` commands. The gui mode is as follows:

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

选择键盘方向键选择'9'并回车，即可进入使用说明:

----

Select the 9 key and press Enter to enter the usage instructions：

```
Welcome to buildenv ().
---------------------------------------
This is a simple pkg-manager for C/C++.

1. How to use it to build cmake project: 
option1: set(CMAKE_TOOLCHAIN_FILE "/mnt/data/work_phil/Golang/buildenv/scripts/toolchain_file.cmake")
option2: cmake .. -DCMAKE_TOOLCHAIN_FILE=/mnt/data/work_phil/Golang/buildenvs/script/toolchain_file.cmake

2. How to use it to build makefile project: 
source /mnt/data/work_phil/Golang/buildenv/script/environment

[ctrl+c/q -> quit]
```

关于cli模式的使用，请参考以下文章:

----

About to to use buildenv in cli mode please read docs below:

1. [buildenv是如何工作的 ----------------- how it works](./docs/01_how_it_works.md)
2. [如何初始化buildenv ------------------- how to init buildenv](./docs/02_init_buildenv.md)
3. [如何添加一个新的平台 ----------------- how to add new platform](./docs/03_add_new_platform.md)
4. [如何添加一个新的项目 ----------------- how to add new project](./docs/04_add_new_project.md)
5. [如何添加一个新的工具 ----------------- how to add new tool](./docs/05_add_new_tool.md)
6. [如何添加一个新的三方库 --------------- how to add new port](./docs/06_add_new_port.md)
7. [如何选择一个平台作为当前平台 -------- how to select platform](./docs/07_how_to_select_platform.md)
8. [如何选择一个项目作为当前项目 -------- how to select project](./docs/08_how_to_select_project.md)
9. [如何集成buildenv --------------------- how to integrate buildenv](./docs/09_integrate_buildenv.md)
10. [如何安装一个三方库 ------------------- how to install a port](./docs/10_how_to_install_port.md)
11. [如何卸载一个三方库 ------------------- how to uninstall a port](./docs/11_how_to_uninstall_port.md)
12. [如何生成cmake配置文件 --------------- how to generate cmake config files](./docs/12_how_to_generate_cmake_config.md)
13. [如何共享安装的三方库 ----------------- how to share installed packages](./docs/13_how_to_share_installed_libraries.md)

## 如何参与贡献 - How to Contribute.

1.  Fork 本仓库
2.  新建 feature_xxx 分支
3.  提交代码
4.  新建 Pull Request

---

1.  Fork this repository.
2.  Create a new branch feature_xxx.
3.  Submit your code changes.
4.  Create a Pull Request.