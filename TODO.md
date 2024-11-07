- 不是所有的编译工具都是必填的，需要支持读field上的自定义tag：

```json
"env_vars": {
    "CC": "aarch64-none-linux-gnu-gcc",
    "CXX": "aarch64-none-linux-gnu-g++",
    "FC": "aarch64-none-linux-gnu-gfortran",
    "RANLIB": "aarch64-none-linux-gnu-ranlib",
    "AR": " aarch64-none-linux-gnu-ar",
    "LD": "aarch64-none-linux-gnu-ld",
    "NM": "aarch64-none-linux-gnu-nm",
    "OBJDUMP": "aarch64-none-linux-gnu-objdump",
    "STRIP": "aarch64-none-linux-gnu-strip"
}
```

- 通过获取资源文件的大小和本地文件大小对比，来判断是否需要重新下载覆盖
- platform添加native的支持
- 提供quiet cli参数，用于在toolchain调用buildenv时候不输出过程信息
- cli未提供create_platform的支持
- go内部拼装PATH环境路径，估计是有问题的，可能不该这么用