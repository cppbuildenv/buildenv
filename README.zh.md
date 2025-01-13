# BuildEnv

英文版本的README, 请查看 [README.md](./README.md).

## 介绍

这是一个用 **Go语言** 实现的 **C/C++ 包管理器**，不需要掌握额外的脚本语言，只需了解 **JSON** 格式即可轻松管理包。该包管理器基于 **CMake**，作为 **CMake** 的补充，主要解决 **CMake 在多架构平台交叉编译环境下的编译、包管理以及所属工具资源绑定问题**。

## 背景问题

CMake长期以来仅提供了 `find_package`和`find_program` 等功能，但缺乏对包的管理能力，特别是在以下几个方面：

1. 编译所需要的工具获取和环境配置，如toolchain, rootfs, cmake，nasm等需要手动安装和配置环境变量；
2. 三方库编译后安装目录和依赖库寻找目录缺乏统一管理，需要手动配置；
3. 交叉编译支持方面，CMake允许通过指定CMAKE_TOOLCHAIN_FILE来配置交叉编译环境，但仍需手动配置。

## 为什么不使用现有的包管理工具？

尽管 Conan和Vcpkg等第三方包管理工具在社区中已经得到了广泛使用，但它们并不能完全满足某些需求：

- **Conan**：虽然功能强大，但依赖于额外的 **Python** 语言，且上手成本较高。因为 Conan 不仅支持 **CMake**，还支持 **Meson**、**Makefile**、**MSBuild**、**SScon**、**QMake**、**Bazaar** 等构建系统，这导致其 API 封装较深，需要更多时间学习和上手，对于本来**CMake**掌握就一般的同学无疑又增加了额外新的学习成本。
  
- **Vcpkg**：相对容易上手，但由于 **国内网络环境问题**使用体验较差，几乎无法正常使用，而且**Vcpkg**对于三方库的版本管理默认仅管理最新版本，这让项目里依赖三方库的特定版本管理不是那么直接。

另外，**Conan** 和 **Vcpkg** 都未能有效管理 **交叉编译环境**，在多个平台的交叉编译时，开发者通常需要手动配置 toolchain 和 rootfs, 以及各种tool的配置，这样不仅繁琐，而且容易出错。

## 解决方案：

为了解决上述问题，**buildenv** 作为一个新的工具应运而生，主要解决以下几个核心问题：

1. **支持管理三方库的安装目录以及编译期间依赖库的寻找目录**：
    - 给CMake项目全局设置 CMAKE_PREFIX_PATH, CMAKE_INSTALL_PREFIX；
    - 给Unix Makefiles项目全局设置 --prefix；
    - 让Unix Makefiles项目在编译期间能通过pc文件找到子依赖，即便当前workspace目录迁移了；

2. **支持自动管理编译工具**：  
通过配置实现自动下载 `toolchain`、`sysroot` 和 `CMake` 等以及配置其环境变量；

3. **支持生成CMake配置文件**：  
对于非CMake作为构建工具的三方库，可以自动生成对应的cmake config文件，方便在CMake项目中使用；

4. **支持指定三方库的install和uninstall**:  
自动编译和安装子依赖，支持卸载库同时卸载子依赖；

5. **支持编译缓存共享**:  
通过配置`cache_dirs`，可进行局域网内网盘来托管和读取`install文件缓存`；

有关更多详细信息，请参阅Docs。

## 安装教程

下载`golang sdk`，然后直接`go build`，即可编译成功。

## 使用说明

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

选择键盘方向键选择'9'并回车，即可进入[使用说明](./docs/01_how_it_works.md).：

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

## 如何参与贡献

1.  Fork 本仓库
2.  新建 Feat_xxx 分支
3.  提交代码
4.  新建 Pull Request