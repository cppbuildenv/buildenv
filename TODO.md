问题    | 状态
-------| -----
不是所有的编译工具都是必填的，需要支持读field上的自定义tag(改成了判空，空则不校验，但CC和CXX是默认必须校验的)  | ✔
通过获取资源文件的大小和本地文件大小对比，来判断是否需要重新下载覆盖  | ✔
platform添加native的支持  | ✔
提供silent cli参数，用于在toolchain调用buildenv时候不输出过程信息  | ✔
cli未提供create_platform的支持  | ✔
go内部拼装PATH环境路径，估计是有问题的，可能不该这么用  | ✔
toolchain里定义CC和CXX时候后面拼上 "--sysroot=xxx"  | ✔
三方库如果编译后产生pkg-config文件，需要加入系统变量（ffmpeg在通过pkg-config寻找libx265）  | ✔
path不存在，应该先尝试用下载后的文件解压，而不是直接重复下载，重新下载的前提是md5等校验没通过  | ✔
每个installed的库添加platform目录，目的是为了支持不同平台的库共存  | ✔
执行-setup打印所有已经准备好的tool和已经安装的port  | ✔
环境变量用os.PathListSeparator拼接  | ✔
将download的资源统一解压到内部的tools目录  | ✔
拓展project，将packages的配置文件放到project目录下  | ✔
menu cli的实现可以考虑用面向对象思维简化  | ✔
git在下载代码时候没有过程log  | ✔
添加-install参数，用于指定三方库的编译  | ✔
--sysroot和--cross-prefix自动设置  | ✔
git 同步代码需要优化  | ✔
预编译好的三方库需要支持uninstall  | ✔
支持uninstall功能, 同时支持recursive 模式  | ✔
makefile的安装路径和依赖寻找路径应该自动管理 | ✔
install 三方库的时候，如果已经配置到project里了，无需指定版本  | ✔
cmd/cli缺少创建和选择project的功能  | ✔
一个项目配置同名不同版本的port是禁止的  | ✔
支持编译库为native的  | ✔
usage 里的颜色需要优化  | ✔
有的toolchain或者tool不是绿色版，不能托管到buildenv里，需要绝对路径指向  | ✔
在project中支持配置cmake变量和C++宏  | ✔
makefile编译前不支持配置环境变量，例如：export CFLAGS="-mfpu=neon"  | ✔
三方库以目录方式维护，内部放不同版本的配置  | ✔
终端输出实现需要再简化  | ✔
当tool不存在，在执行install的时候不会触发下载  | ✔
支持打patch  | ✔
第一次使用交互需要优化  | ✔
cmake_config的配置独立于version文件之外  | ✔
有些pc文件产生做share目录，而不是lib目录，需要统一移动到lib目录（libz）  | ✔
通过一个中间临时目录来实现收集install的文件清单  | ✔
支持通过命令创建tool和port  | ✔
支持编译缓存共享  | ✔
内部出现同一个库的不同版本依赖情况给与报错提示 | ✔
支持clone时候连同submodule一起clone  | ✔
支持meson  |  ✔
支持ninja  |  ✔
将buildtype抽象到各个buildsystem里 | ✔
支持autotools  | ✔
运行tools需要将内部lib路径加入到LD_LIBRARY_PATH  | ✘
支持 buildenv -upgrade 升级  | ✘
动态生成的cmake config文件（windows还没测试）| ✘
支持windows下工作  | ✘
下载的库暂不支持生成cmake config文件  | ✘
在创建的新tool和port里添加注释  | ✘
支持ccache  | ✘
支持fork到私有仓库  | ✘
支持在project里覆盖默认port的配置  | ✘
如果发现资源包size跟最新不匹配，即便已经解压了也要重新下载 | ✘
支持export功能 | ✘
支持在project里定义CMAKE_CXX_FLAGS和CMAKE_C_FLAGS，以及LDFLAGS | ✘
检测代码如果跟目标不匹配, 什么都不做，同时提供sync命令用于强行同步代码 | ✘
校验是否真的installed还需要判断文件是否存在 | ✘