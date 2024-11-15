# buildenv

&emsp;&emsp;这是一个用Go语言实现的C/C++的包管理器，无需掌握额外的脚本语言，只要懂JSON就可以轻松的管理自己的包管理器，它依托于CMake，只作为对CMake的补充，主要解决CMake多芯片平台交叉编译环境下的包管理和工具资源下载的问题。  
&emsp;&emsp;向来CMake只提供了find_package的能力，即：包寻找能力。但三方库编译后安装目录、依赖库寻找目录却没有统一管理，即：包的管理能力是缺失的。  
&emsp;&emsp;C/C++虽然缺少一个官方的包管理，但不缺乏一些第三方的包管理，比较流行的第三方包管理有Conan和Vcpkg，为什么还要创造一个新的呢？主要原因是conan依赖额外的Python语言，而且上手成本较高，主要是因为Conan的出现不仅仅针对CMake，为了同时支持CMake、Meson、Makefile、MSBuild、SScon、QMake、Bazaar等作为你的项目的构建系统而封装的很深，需要更多的时间学习它封装后的Python API。vcpkg上手倒是比较容易，但是国内的网络环境下vcpkg几乎是无法正常使用的。  
&emsp;&emsp;其实，不管是conan还是vcpkg，都缺失了对交叉编译环境的管理支持，对于多平台的交叉编译环境管理，往往需要自己手动放置toolchain和rootfs到本地目录，然后再手动写一个`toolchain.cmake`, 然后将`toolchain`和`root`路径配置到`toolchain.cmake`里，这个过程对于维护多个平台的工程项目来说比较繁琐，还容易出错。又或者把多个平台的`toolchain`、`rootfs`、`tools`等全部提交到一个代码仓库并提供写死路径的`toolchain.cmake`, 时间久了仓库十分臃肿，也不是个好办法。

因此，`buildenv`的出现就是为了解决上面的2个问题：

1. 管理三方库编译安装到固定的目录，然后此目录也是别的三方库寻找依赖的目录；
2. 自动下载编译工具（`toolchian`、`sysroot`、`cmake`等tool）

>上面只是buildenv解决的核心问题，但`buildenv`还有其他刚需功能，比如：自动生成cmake用于交叉编译的`toochain`文件、导出依赖库等, 详细请阅读Wiki。


#### 安装教程

下载`golang sdk`，然后直接`go build`，即可编译成功。

#### 使用说明

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

#### 参与贡献

1.  Fork 本仓库
2.  新建 Feat_xxx 分支
3.  提交代码
4.  新建 Pull Request