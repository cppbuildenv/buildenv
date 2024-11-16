# buildenv的workspace目录结构

```json
├── buildenv.json
├── buildtrees
│   ├── ffmpeg
│   │   ├── aarch64-linux-jetson-nano-Release
│   │   ├── aarch64-linux-jetson-nano-Release-build.log
│   │   ├── aarch64-linux-jetson-nano-Release-clone.log
│   │   ├── aarch64-linux-jetson-nano-Release-configure.log
│   │   ├── aarch64-linux-jetson-nano-Release-install.log
│   │   └── src
│   ├── gflags
│   ├── x264
│   └── x265
├── conf
│   ├── platforms
│   │   ├── aarch64-linux-jetson-nano.json
│   │   ├── native.json
│   │   └── x86_64-linux-ubuntu-22.04.5-base.json
│   ├── ports
│   │   ├── ffmpeg-v3.4.json
│   │   ├── gflags-v2.2.2.json
│   │   ├── glog-v0.6.0.json
│   │   ├── qt5-v5.15-x86_64-linux.json
│   │   ├── x264-statble.json
│   │   └── x265-3.4.json
│   ├── README.md
│   └── tools
│       ├── cmake-3.30.5-linux-x86_64.json
│       └── nasm-2.16.03.json
├── downloads
│   ├── cmake-3.30.5-linux-x86_64/
│   ├── cmake-3.30.5-linux-x86_64.tar.gz
│   ├── gcc-arm-9.2-2019.12-x86_64-aarch64-none-linux-gnu/
│   ├── gcc-arm-9.2-2019.12-x86_64-aarch64-none-linux-gnu.tar.gz
│   ├── nasm-2.16.03-x86_64-linux/
│   ├── nasm-2.16.03-x86_64-linux.tar.gz
│   ├── jetson-nano-sd-card-image-r32.x.x_aarch64/
│   └── jetson-nano-sd-card-image-r32.x.x_aarch64.tar.gz
├── installed
│   ├── aarch64-linux-jetson-nano-Release
│   │   ├── bin
│   │   ├── include
│   │   ├── lib
│   │   └── share
│   └── buildenv
│       └── aarch64-linux-jetson-nano-Release.list
└── script
   	  ├── buildenv.cmake
      └── buildenv.sh
```

目录结构说明
----------------

- buildenv.json: workspace的配置文件
- buildtrees: 构建的中间产物目录
- conf: 配置文件目录
- downloads：toolchain、rootfs以及tools的下载和解压目录
- installed：所有port的安装目录
- script: 脚本目录


# 如何创建workspaced的配置文件
 
&emsp;&emsp;**buildenv**的工作依赖一套配置，在配置中描述采用了什么**toolchain**、**rootfs**、**cmake**、**tool**以及依赖哪些三方库，然后**buildenv**会根据配置去下载资源、拉取代码、编译构建工具、安装到指定目录，此文件就是`buildenv.json`。

## 1. 创建配置文件

buildenv提供了两种命令行交互，一种是通过命令行参数完成创建配置文件，另一种是通过交互式cli完成创建配置文件。

### 1.1 命令行参数创建

第一次执行`./buildenv -sync`会在当前`workspace`目录下产生一个全局配置文件，即：`buildenv.json`。

```shell
$ ./buildenv -sync
[✔] ======== buildenv.json is created but need to config it later.
```

### 1.2 交互式cli创建

```
$ ./buildenv -ui

    Please select one from the menu...                     
                                                           
  > 1. Init or sync buildenv's config repo.                
    2. Create a new platform, but need to config it later.
    3. Select a platform as your build target platform.    
    4. Install buildenv.                                   
    5. About and Usage.                                    
                                                           
                                                           
    ↑/k up • ↓/j down • q quit • ? more                    
```

>通过键盘上下键选择，然后回车即可进入创建`buildenv.json`的配置.

填写完整的`buildenv.json`如下：

```json
{
    "platform": "aarch64-linux-jetson-nano",
    "conf_repo": "ssh://git@192.168.xxx.xxx:xxx/buildenv_conf.git",
    "conf_repo_ref": "master",
    "job_num": 32
}
```

>对于一个完整的项目，配置项目很多，包含不同的平台里的对应**toolchain**、**rootfs**、**tool**以及三方库的详细编译配置，因此我们推荐将它们的配置放在一个单独的仓库中，然后通过**conf_repo**和**conf_repo_ref**来引用。  
>所以，只要创建一个空的git仓库，然后把**repo url**和**ref(分支名或者tag名)** 填进去即可（此时，**platform** 是空白，先不用管，因为还没有定义**platform**）。

## 2. 定义配置文件仓库文件结构

```
conf
  ├── platforms
  ├── ports
  └── tools
```

>确保拥有上面的文件结构，后续在**platforms**、**ports**和**tools**中创建对应的配置文件，参考下面的实例。

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

### 3. 同步配置文件

当存放配置文件的git仓库创建好，甚至创建了**platforms**、**ports**、**tools**并提交到了git仓库，随后第二次执行`./buildenv -sync`命令，即可同步配置文件到当前**workspace**的**conf**目录。

依然，我们有两种方式可以同步配置文件：

1. 命令行参数同步

```json
$ ./buildenv -sync
HEAD is now at 2b700c7 update config to support multi build configs
Already on 'develop'
Your branch is up to date with 'origin/develop'.
Already up to date.

[✔] ======== conf repo is synchronized.
```

2. 交互式cli同步

```json
$ ./buildenv -ui

    Please choose one from the menu...                     
                                                           
  > 1. Init or sync buildenv's config repo.                
    2. Create a new platform, it requires completion later.
    3. Choose a platform as your build target platform.    
    4. Install buildenv.                                   
    5. About and Usage.                                    
                                                           
    ↑/k up • ↓/j down • q quit • ? more 
```

## 3. 管理platform、tool以及port

1. [如何配置platform](./platform.md)
2. [如何配置tool](./tool.md)
3. [如何配置port](./port.md)

## 4. 安装buildenv

TODO: write later...

## 5. 关于和使用

TODO: write later...