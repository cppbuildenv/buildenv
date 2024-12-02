问题    | 状态
-------| -----
不是所有的编译工具都是必填的，需要支持读field上的自定义tag(改成了判空，空则不校验，但CC和CXX是默认必须校验的)  | ✔
通过获取资源文件的大小和本地文件大小对比，来判断是否需要重新下载覆盖  | ✔
platform添加native的支持  | ✔
提供silent cli参数，用于在toolchain调用buildenv时候不输出过程信息  | ✔
cli未提供create_platform的支持  | ✔
go内部拼装PATH环境路径，估计是有问题的，可能不该这么用  | ✔
支持 buildenv -upgrade 升级  | ✘
toolchain里定义CC和CXX时候后面拼上 "--sysroot=xxx"  | ✔
运行tools需要将内部lib路径加入到LD_LIBRARY_PATH  | ✘
三方库如果编译后产生pkg-config文件，需要加入系统变量（ffmpeg在通过pkg-config寻找libx265）  | ✔
cli里添加命令：编译指定的某个port，最好以列表方式呈现，让用户选择  | ✘
有的toolchain或者tool不是绿色版，不能托管到buildenv里，需要绝对路径指向  | ✘
path不存在，应该先尝试用下载后的文件解压，而不是直接重复下载，重新下载的前提是md5等校验没通过  | ✔
每个installed的库添加platform目录，目的是为了支持不同平台的库共存  | ✔
执行-verify打印所有已经准备好的tool和已经安装的port  | ✔
makefile编译前不支持配置环境变量，例如：export CFLAGS="-mfpu=neon"  | ✘
环境变量用os.PathListSeparator拼接  | ✔
添加-build参数，用于指定三方库的编译  | ✘
添加offline选项，用于指定不下载  | ✘
将download的资源统一解压到内部的tools目录  | ✔
完事动态生成的cmake config文件（静态库还没支持）| ✘
支持windows下工作  | ✘