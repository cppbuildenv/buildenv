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