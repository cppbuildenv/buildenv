# 如何配置platfomr

`platform`的配置是 `buildenv`编译平台目标，它定义了这个`platform`所需要的`toolchain`、`rootfs`、`tools`以及`ports`。

## 1. 创建配置文件

执行`./buildenv -create_platform aarch64-linux-jetson-nano`会自动创建一个platform文件，即：`conf/platforms/aarch64-linux-jetson-nano.json`。

```shell
$ buildenv -create_platform aarch64-linux-jetson-nano
[✔] ======== aarch64-linux-jetson-nano is created but need to config it later.
```

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
        "env_vars": {
            "TOOLCHAIN_PREFIX": "",# ------ 编译工具前缀， 如：aarch64-linux-gnu-
            "CC": "",              # ------ 编译工具， 如：aarch64-linux-gnu-gcc
            "CXX": "",             # ------ 编译工具， 如：aarch64-linux-gnu-g++
            "FC": "",
            "RANLIB": "",
            "AR": "",
            "LD": "",
            "NM": "",
            "OBJDUMP": "",
            "STRIP": ""
        }
    },
    "tools": [],                   # ------ 编译工具， 如：cmake, make, nasm等
    "packages": []                 # ------ 三方库， 如：gflags, opencv, qt5, ffmpeg等
}
```

>对于交叉编译，`rootfs`和`toolchain`是必须的，`tools`和`packages`是可选的，但如果想要指定版本的CMake需要把cmake配置到tools里。
>一般`toolchain`里`CC`和`CXX`是必须设置的，其它根据项目需要来配置。

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
        "env_vars": {
            "TOOLCHAIN_PREFIX":"aarch64-none-linux-gnu-",
            "CC": "aarch64-none-linux-gnu-gcc",
            "CXX": "aarch64-none-linux-gnu-g++",
            "FC": "aarch64-none-linux-gnu-gfortran",
            "RANLIB": "aarch64-none-linux-gnu-ranlib",
            "AR": "aarch64-none-linux-gnu-ar",
            "LD": "aarch64-none-linux-gnu-ld",
            "NM": "aarch64-none-linux-gnu-nm",
            "OBJDUMP": "aarch64-none-linux-gnu-objdump",
            "STRIP": "aarch64-none-linux-gnu-strip"
        }
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

[下一步：如何配置配置tool.](./tool.md)