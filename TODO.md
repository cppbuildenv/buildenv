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
执行-verify打印所有已经准备好的tool和已经安装的port  | ✔
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
运行tools需要将内部lib路径加入到LD_LIBRARY_PATH  | ✘
支持 buildenv -upgrade 升级  | ✘
有的toolchain或者tool不是绿色版，不能托管到buildenv里，需要绝对路径指向  | ✘
makefile编译前不支持配置环境变量，例如：export CFLAGS="-mfpu=neon"  | ✘
动态生成的cmake config文件（windows还没测试）| ✘
支持windows下工作  | ✘
usage 里的颜色需要优化  | ✘
下载的库暂不支持生成cmake config文件  | ✘
支持通过命令创建tool和port  | ✘
在创建的新tool和port里添加注释  | ✘
在project中支持配置cmake变量和C++宏  | ✘
支持autotools  | ✘
支持meson  | ✘
支持ninja  | ✘