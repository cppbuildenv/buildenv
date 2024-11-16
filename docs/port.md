# 如何配置port

所有的三方库的源码通过定义在**ports**目录里的配置文件来下载，并同时定义如何编译它们。

## 1. 创建port配置文件

port的配置文件模板如下：

```json
{
    "url": "",                  # ------ 源码下载地址（通过后缀识别是git clone还是http下载）
    "version": "",              # ------ 源码版本， 如：stable, master, v1.2.3等
    "dependencies": [],         # ------ 依赖的port， 如：gflags-v2.2.2等
    "build_configs": [          # ------ 这是一个数组，可以配置多个不同平台的编译
        {
            "pattern": "",      # ------ 匹配的平台， 如：*linux*等
            "build_tool": "",   # ------ 编译工具， 如：cmake, make, nasm等
            "arguments": []     # ------ 编译参数， 如：--enable-shared等, 也可以是CMAKE的-D参数等
        }
    ]
}
```

**例如**：

1. 在`conf/ports`目录下创建用makefile构建的三方库的port：即：`ffmpeg-v3.4.json`文件：

```json
{
    "url": "ssh://git@192.168.xxx.xxx:xxx/ffmpeg.git",
    "version": "n4.4",
    "dependencies": [
        "x264-statble"
    ],
    "build_configs": [
        {
            "pattern": "*linux*",
            "build_tool": "make",
            "arguments": [
                "--enable-shared",
                "--disable-static",
                "--disable-x86asm",
                "--disable-programs",
                "--disable-doc",
                "--enable-libx264",
                "--disable-libx265",
                "--enable-gpl",
                "--cross-prefix=${TOOLCHAIN_PREFIX}",
                "--extra-cflags=-I${INSTALLED_DIR}/include",
                "--extra-ldflags=-L${INSTALLED_DIR}/lib",
                "--arch=aarch64",
                "--target-os=linux"
            ]
        }
    ]
}
```

**注意：**

    - dependencies里的配置必须是ports里已经存在的port名字；
    - 由于makefile项目的configure参数非常不统一性，所以这里使用了`arguments`来配置，其中的`${TOOLCHAIN_PREFIX}`和`${INSTALLED_DIR}`会被替换成对应的toolchain和rootfs的路径。

2. 在`conf/ports`目录下创建用cmake构建的三方库的port: `gflags-v2.2.2.json`：

```json
{
    "url": "ssh://git@192.168.xxx.xxx:xxx/gflags.git",
    "version": "v2.2.2",
    "dependencies": [],
    "build_configs": [
        {
            "pattern": "*linux*",
            "build_tool": "cmake",
            "arguments": [
                "-DBUILD_SHARED_LIBS=ON",
                "-DBUILD_STATIC_LIBS=OFF",
                "-DBUILD_TESTING=OFF",
                "-DCMAKE_BUILD_TYPE=Release"
            ]
        }
    ]
}
```

>注意：在cmake构建配置中允许写死的参数，如：`-DCMAKE_BUILD_TYPE=Release`，目的是为了即便当前的platform的`build_type`为`Debug`时，也只编译Release版本，至于`CMAKE_INSTALL_PREFIX`和`CMAKE_PREFIX_PATH`不用指定，它们会由`buildenv`统一管理。