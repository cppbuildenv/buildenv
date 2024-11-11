# buildenv

#### 介绍

&emsp;&emsp;这是一个用Go语言实现的C/C++的包管理器，依托于CMake，作为对CMake的一个补充，主要解决CMake在多平台交叉编译环境下的包管理问题。
&emsp;&emsp;向来CMake提供了包寻找包的能力，但是如何提供一个统一的包管理工具，让用户可以方便的管理自己的包，是一个比较大的问题。  
&emsp;&emsp;无需掌握额外的脚本语言，只要懂JSON就可以轻松的管理自己的包。
&emsp;&emsp;C/C++一直缺少一个官方的包管理，虽然比较流行的第三方包管理有conan和vcpkg，为什么还要搞一个新的呢？主要原因是conan严重依赖python脚本，而且及不容易上手，vcpkg上手倒是比较容易，但是国内的网络环境下vcpkg几乎是无法正常使用的。  
&emsp;&emsp;其实，不管是conan还是vcpkg，都缺失了对交叉编译环境的管理支持，对于多平台的交叉编译环境管理，往往需要自己手动放置toolchain和rootfs到本地目录，然后再手动写一个toolchain.cmake, 然后将toolchain和root路径配置到toolchain.cmake里，这个过程对于维护多个平台的工程项目来说比较繁琐，还容易出错。又或者把多个平台的toolchain、rootfs等全部提交到一个代码仓库并提供写死路径的toolchain.cmake, 时间久了仓库十分臃肿。

#### 安装教程

下载`golang sdk`，然后直接`go build`，即可编译成功。

#### 使用说明

1.  xxxx
2.  xxxx
3.  xxxx

#### 参与贡献

1.  Fork 本仓库
2.  新建 Feat_xxx 分支
3.  提交代码
4.  新建 Pull Request