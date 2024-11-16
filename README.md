# BuildEnv

## 介绍

这是一个用 **Go语言** 实现的 **C/C++ 包管理器**，不需要掌握额外的脚本语言，只需了解 **JSON** 格式即可轻松管理包。该包管理器基于 **CMake**，作为 CMake 的补充，主要解决 **CMake 在多芯片平台交叉编译环境下的包管理和工具资源下载问题**。

## 背景问题

**CMake** 长期以来仅提供了 `find_package` 功能，即包寻找能力，但缺乏对包的管理能力，特别是在以下几个方面：

1. **三方库编译后安装目录** 和 **依赖库寻找目录** 缺乏统一管理。
2. 对于 **交叉编译环境**，CMake 没有提供专门的包管理支持。

## 为什么不使用现有的包管理工具？

尽管 **Conan** 和 **Vcpkg** 等第三方包管理工具在社区中已经得到了广泛使用，但它们并不能完全满足某些需求：

- **Conan**：虽然功能强大，但依赖于额外的 **Python** 环境，且上手成本较高。因为 Conan 不仅支持 **CMake**，还支持 **Meson**、**Makefile**、**MSBuild**、**SScon**、**QMake**、**Bazaar** 等构建系统，这导致其 API 封装较深，需要更多时间学习和上手。
  
- **Vcpkg**：相对容易上手，但由于 **国内网络环境问题**，Vcpkg 的使用体验较差，几乎无法正常使用。

另外，**Conan** 和 **Vcpkg** 都未能有效支持 **交叉编译环境**，在多个平台的交叉编译时，开发者通常需要手动处理 toolchain 和 rootfs 的配置，这样不仅繁琐，而且容易出错。

## 解决方案：`buildenv`

为了解决上述问题，**buildenv** 作为一个新的工具应运而生，主要解决以下两个核心问题：

1. **管理三方库的安装目录**，并提供统一的 **依赖库寻找目录**，使得包管理更为简洁。
2. **自动下载编译工具**，包括 **toolchain**、**sysroot** 和 **CMake** 等生成对应的toolchain file文件，极大简化了交叉编译的配置过程。

## 其他核心功能

除了上述功能，**buildenv** 还提供了以下刚需功能：

- 自动生成 CMake 配置文件，支持交叉编译的 **toolchain** 文件。
- 自动导出依赖库，方便管理和使用。

有关更多详细信息，请参阅 **Wiki** 页面。

## 安装教程

下载`golang sdk`，然后直接`go build`，即可编译成功。

## 使用说明

```
Usage of ./buildenv:
  -build_type string
        called by buildenv.cmake to set CMAKE_BUILD_TYPE. (default "Release")
  -create_platform string
        create a new platform
  -install
        install buildenv so that can use it everywhere
  -select_platform string
        select a platform as build target platform
  -silent
        called by buildenv.cmake to run buildenv in silent mode.
  -sync
        create buildenv.json or sync conf repo defined in buildenv.json.
  -ui
        run buildenv in gui mode.
  -verify
        check and repair toolchain, rootfs, tools and packages for current selected platform.
  -version
        print version.
```

>详细说明请看wiki

## 如何参与贡献

1.  Fork 本仓库
2.  新建 Feat_xxx 分支
3.  提交代码
4.  新建 Pull Request