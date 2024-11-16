# 如何配置tool

tool的配置的目标是下载到指定的目录，然后添加到PATH中，以方便项目在编译过程中可以直接调用。

## 1. 创建tool配置文件

tool的配置文件比较简单，模板如下：

```json
{
    "url": "",
    "path": ""
}
```

**例如**：在`conf/tools`目录下创建`cmake-3.30.5-linux-x86_64.json`文件，内容如下：

```json
{
    "url": "http://192.168.xxx.xxx:xxxx/tools/cmake/cmake-3.30.5-linux-x86_64.tar.gz",
    "path": "cmake-3.30.5-linux-x86_64/bin"
}
```
>所有tool会下载到**workspace**的**downloads**目录里，并会也会解压到**downloads**里。