# How to share installed libraries.

All third-party libraries can build locally for our project. But sometimes we want to share them with other projects. For example, we want to use ffmpeg in our project, but we don't want to build it every time, because it takes a lot of time. Luckyly, buildenv can help us to share them.

## 1. Define `cache_dirs` in `buildenv.json`.

You can define the `cache_dirs` as below, you can define multiple cache dirs, and you can define them as readable or writable. If the cache dir is writable, buildenv will try to copy the installed libraries to the cache dir. If the cache dir is readable, buildenv will try to find the installed libraries in the cache dir first.

```
{
    "conf_repo_url": "ssh://git@192.168.0.78:7999/isw/buildenv-conf.git",
    "conf_repo_ref": "master",
    "platform_name": "x86_64-linux-ubuntu-20.04.5",
    "project_name": "project_01",
    "job_num": 32,
    "cache_dirs": [        # ------------------------------ look here
        {
            "dir": "/mnt/buildenv_cache",
            "readable": false,
            "writable": false
        }
    ]
}
```

# 2. Build and install by buildenv from source code.

When a third-party library is compiled and installed from source, its installation files will be packaged and stored in the cache directory, the cache directory will be like this:

```
mnt
└── buildenv_cache
    └── x86_64-linux-ubuntu-20.04
        └── project_01
            └── Release
                ├── ffmpeg@3.4.13.tar.gz
                ├── opencv@4.5.1.tar.gz
                ├── sqlite3@3.49.0.tar.gz
                ├── x264@stable.tar.gz
                ├── x265@4.0.tar.gz
                └── zlib@v1.3.1.tar.gz
```

When type `buildenv -install xxx@yyy`, buildenv will try to find the installed files in the cache directories one by one, if not found, it will build and install it from source code.  
It must satisfy five matching elements to conclude that it is the cache target being searched for.   
**The fine elements are:**

1. platform name: for example `x86_64-linux-ubuntu-20.04.5`
2. project name: for example `project_01`
3. library name: for example `ffmpeg`
4. library version: for example `3.4.13`
5. build config: for example `Release`