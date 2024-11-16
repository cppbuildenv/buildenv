# 如何配置BuildEnv

`C/C++`的编译往往依赖依赖`toolchain`、`rootfs`以及构建工具，即：`CMake`、`Make`、`QMake`等，因为这里主角是`BuildEnv`，它是依托`CMake`且是`CMake`的一个补充，因此使用`BuildEnv`必须提供`CMake`。

`BuildEnv`的工作依赖一套配置，在配置中描述你采用了什么`toolchain`、`rootfs`、`cmake`、`tool`以及依赖哪些三方库，然后`BuildEnv`会根据配置去下载资源、拉取代码、编译构建工具、安装到指定目录。

## 1. 创建配置文件

执行`./buildenv -sync`会产生一个配置模板文件，即：`buildenv.json`。

```shell
$ ./buildenv -sync
[✔] ======== buildenv.json is created but need to config it later.
```

填写完整的`buildenv.json`如下：

```json
{
    "platform": "aarch64-linux-d100-j721e",
    "conf_repo": "ssh://git@192.168.xxx.xxx:xxx/buildenv_conf.git",
    "conf_repo_ref": "master",
    "job_num": 32
}
```

>对于一个完整的项目，配置项目很多，包含不同的平台里的对应`toolchain`、`rootfs`、`tool`以及三方库的详细编译配置，因此我们推荐将它们的配置放在一个单独的仓库中，然后通过`conf_repo`和`conf_repo_ref`来引用。  
>所以，只要创建一个空的git仓库，然后把`repo url`和`ref(分支名或者tag名)`填进去即可（此时，`platform`是空白，先不用管，因为还没有定义`platform`）。

## 2. 定义配置文件仓库文件结构

```
conf
  ├── platforms
  ├── ports
  └── tools
```

>确保拥有上面的文件结构，后续在`platforms`、`ports`和`tools`中创建对应的配置文件，参考下面的实例。

```
conf
  ├── platforms
  │   ├── aarch64-linux-jetson-nano.json
  │   ├── native.json
  │   └── x86_64-linux-ubuntu-22.04.5-base.json
  ├── ports
  │   ├── ffmpeg-v3.4.json
  │   ├── gflags-v2.2.2.json
  │   ├── glog-v0.6.0.json
  │   ├── qt5-v5.15-x86_64-linux.json
  │   ├── x264-statble.json
  │   └── x265-3.4.json
  └── tools
      ├── cmake-3.30.5-linux-x86_64.json
      └── nasm-2.16.03-linux-x86_64.json
```

[下一步：如何配置配置platform.](./platform.md)