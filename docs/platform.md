# 如何配置platfomr

**platform**的配置文件存放在**conf/platforms**下，在此文件中定义了这个`platform`所需要的`toolchain`、`rootfs`、`tools`以及`ports`。

## 1. 创建platform配置文件

### 1.1 命令行参数创建

```shell
$ ./buildenv -create_platform aarch64-linux-jetson-nano
[✔] ======== aarch64-linux-jetson-nano is created but need to config it later.
```

>一个platform文件被创建，即：`conf/platforms/aarch64-linux-jetson-nano.json`。

### 1.2 交互式cli创建

```
$ ./buildenv -ui

    Please select one from the menu...                     
                                                           
    1. Init or sync buildenv's config repo.                
  > 2. Create a new platform, it requires completion later.
    3. Select a platform as your build target platform.    
    4. Install buildenv.                                   
    5. About and Usage.                                    
                                                           
                                                           
    ↑/k up • ↓/j down • q quit • ? more                    
```

通过键盘上下键选择，然后回车即可进入创建platform的配置：

```
Please input your platform name: 

> for example: x86_64-linux-ubuntu-20.04...                                                            

[esc -> back | ctrl+c/q -> quit]
```

>platform名字一定要体现出平台特性，如：`aarch64-linux-jetson-nano`、`x86_64-linux-ubuntu-20.04`等。

创建的`aarch64-linux-jetson-nano.json`内容如下：

```json
{
    "rootfs": {
        "url": "",                 # ----- rootfs的下载地址
        "path": "",                # ----- rootfs解压后暴露到PATH里的路径
        "pkg_config_path": []      # ----- pkg-config的搜索路径，一般是rootfs的usr/lib/pkgconfig
    },
    "toolchain": {
        "url": "",                 # ----- toolchain的下载地址
        "path": "",                # ----- toolchain解压后暴露到PATH里的路径
        "system_name": "",         # ----- 目标系统名, 如：linux, darwin, windows
        "system_processor": "",    # ----- 目标系统架构, 如：aarch64, x86_64, i386
        "toolchain_prefix": "",    # ----- 编译工具前缀， 如：aarch64-linux-gnu-
        "cc": "",                  # ----- 编译工具， 如：aarch64-linux-gnu-gcc
        "cxx": "",                 # ----- 编译工具， 如：aarch64-linux-gnu-g++
        "fc": "",
        "randlib": "",
        "ar": "",
        "ld": "",
        "nm": "",
        "objdump": "",
        "strip": ""
    },
    "tools": [],                   # ----- 编译工具， 如：cmake, make, nasm等
    "packages": []                 # ----- 三方库， 如：gflags, opencv, qt5, ffmpeg等
}
```

>对于交叉编译，`rootfs`和`toolchain`是必须的，`tools`和`packages`是可选的，但如果想要指定版本的CMake需要把cmake配置到tools里;
>一般`toolchain`里`cc`和`cxx`是必须设置的，其它根据项目需要来配置;
>`toolchain`和`rootfs`下载到`workspace`的`downloads`目录里，并会也会解压到`downloads`里。

完整参考如下：

```json
{
    "rootfs": {
        "url": "http://192.168.xxx.xxx:xxx/rootfs/jetson-nano-sd-card-image-r32.x.x_aarch64.tar.gz",
        "path": "jetson-nano-sd-card-image-r32.x.x_aarch64",
        "pkg_config_path": [
            "usr/lib/pkgconfig"
        ]
    },
    "toolchain": {
        "url": "http://192.168.xxx.xxx:xxx/toolchain/gcc-arm-9.2-2019.12-x86_64-aarch64-none-linux-gnu.tar.gz",
        "path": "gcc-arm-9.2-2019.12-x86_64-aarch64-none-linux-gnu/bin",
        "system_name": "Linux",
        "system_processor": "aarch64",
        "toolchain_prefix": "aarch64-none-linux-gnu-",
        "cc": "aarch64-none-linux-gnu-gcc",
        "cxx": "aarch64-none-linux-gnu-g++",
        "fc": "aarch64-none-linux-gnu-gfortran",
        "ranlib": "aarch64-none-linux-gnu-ranlib",
        "ar": "aarch64-none-linux-gnu-ar",
        "ld": "aarch64-none-linux-gnu-ld",
        "nm": "aarch64-none-linux-gnu-nm",
        "objdump": "aarch64-none-linux-gnu-objdump",
        "strip": "aarch64-none-linux-gnu-strip"
    },
    "tools": [
        "cmake-3.30.5-linux-x86_64",
        "nasm-2.16.03-linux-x86_64"
    ],
    "packages": [
        "gflags-v2.2.2",
        "ffmpeg-v3.4"
    ]
}
```

如果是用于本地编译的platform，`rootfs`和`toolchain`可以不配置，如下:

```json
{
    "tools": [
        "cmake-3.30.5-linux-x86_64"
    ],
    "dependencies": [
        "gflags-v2.2.2",
        "ffmpeg-v3.4"
    ]
}
```