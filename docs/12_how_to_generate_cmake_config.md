# How to generate cmake config.

As we all known, a lot of 3rd party libs are not built with cmake, after installed, they would not generate cmake config files, which would be used by cmake to find them. Although we can use `pkg-config` to find them, but only can be used in linux. Now, buildenv can generate cmake config files for them, which can be used in any platform.

## 1. Library without components.

For example, x264, you should create a [version]@cmake_config.json file in the same directory as the [version].json file.

```
└── x264
    ├── stable@cmake_config.json
    └── stable.json
```

The file content would be like this, we can define different file names for different platforms.

```
{
    "namespace": "x264",
    "linux_static": {
        "filename": "libx264.a"
    },
    "linux_shared": {
        "filename": "libx264.so.164",
        "soname": "libx264.so"
    },
    "windows_static": {
        "filename": "x264.lib"
    },
    "windows_shared": {
        "filename": "x264.dll",
        "impname": "x264.lib"
    }
}
```

Then you can define the cmake config in the [version].json file for different build_config sections.

```
{
    "url": "https://gitee.com/chooosky/x264.git",
    "name": "x264",
    "version": "stable",
    "build_configs": [
        {
            "pattern": "x86_64-linux*",
            "build_tool": "makefiles",
            "library_type": "shared",
            "env_vars": [ ],
            "arguments": [
                "--host=${HOST}",
                "--sysroot=${SYSROOT}",
                "--cross-prefix=${CROSS_PREFIX}",
                "--disable-cli",
                "--enable-pic"
            ],
            "dependencies": [ ],
            "cmake_config": "linux_shared"
        },
        {
            "pattern": "aarch64-linux*",
            "build_tool": "makefiles",
            "library_type": "static",
            "env_vars": [ ],
            "arguments": [
                "--host=${HOST}",
                "--sysroot=${SYSROOT}",
                "--cross-prefix=${CROSS_PREFIX}",
                "--disable-cli",
                "--enable-pic"
            ],
            "dependencies": [ ],
            "cmake_config": "linux_static"        # ------ look here !!!
        }
    ]
}
```

After building and installing, you can see the generated cmake config files as below:

```
lib
└── cmake
    └─── x264
        ├── x264Config.cmake
        ├── x264ConfigVersion.cmake
        ├── x264Targets.cmake
        └── x264Targets-release.cmake
```

Finally, you can use it in your cmake project as below:

```cmake
find_package(x264 REQUIRED)
target_link_libraries(${PROJECT_NAME} PRIVATE x264::x264)
```

> Please note that, the namespace is defined in cmake_config file, if not defined, it would be the same as the library name.

## 2. Library with components.

Alike library without components, you still need to define a [version]@cmake_config.json file in the same directory as the [version].json file.

```
└── ffmpeg
    ├── n3.4.13@cmake_config.json
    └── n3.4.13.json
```

The file content would be like this, it would be a little different from library without components, we can define different file names for different platforms and components.

```
{
    "namespace": "FFmpeg",
    "linux_shared": {
        "components": [
            {
                "component": "avutil",
                "soname": "libavutil.so.55",
                "filename": "libavutil.so.55.78.100",
                "dependencies": [ ]
            },
            {
                "component": "avcodec",
                "soname": "libavcodec.so.57",
                "filename": "libavcodec.so.57.107.100",
                "dependencies": [
                    "avutil"
                ]
            },
            {
                "component": "avdevice",
                "soname": "libavdevice.so.57",
                "filename": "libavdevice.so.57.10.100",
                "dependencies": [
                    "avformat",
                    "avutil"
                ]
            },
            {
                "component": "avfilter",
                "soname": "libavfilter.so.6",
                "filename": "libavfilter.so.6.107.100",
                "dependencies": [
                    "swscale",
                    "swresample"
                ]
            },
            {
                "component": "avformat",
                "soname": "libavformat.so.57",
                "filename": "libavformat.so.57.83.100",
                "dependencies": [
                    "avcodec",
                    "avutil"
                ]
            },
            {
                "component": "postproc",
                "soname": "libpostproc.so.54",
                "filename": "libpostproc.so.54.7.100",
                "dependencies": [
                    "avcodec",
                    "swscale",
                    "avutil"
                ]
            },
            {
                "component": "swresample",
                "soname": "libswresample.so.2",
                "filename": "libswresample.so.2.9.100",
                "dependencies": [
                    "avcodec",
                    "swscale",
                    "avutil",
                    "avformat"
                ]
            },
            {
                "component": "swscale",
                "soname": "libswscale.so.4",
                "filename": "libswscale.so.4.8.100",
                "dependencies": [
                    "avcodec",
                    "avutil",
                    "avformat"
                ]
            }
        ]
    }
}
```

> Different components may have different dependencies, so we need to define them in the `dependencies` field.


The next step is to define the cmake config in the [version].json file for different build_config sections.

```
{
    "url": "ssh://git@192.168.12.18:7999/thdpty/ffmpeg.git",
    "name": "ffmpeg",
    "version": "n3.4.13",
    "build_configs": [
        {
            "pattern": "*linux*",
            "build_tool": "makefiles",
            "library_type": "shared",
            "env_vars": [ ],
            "arguments": [
                "--sysroot=${SYSROOT}",
                "--cross-prefix=${CROSS_PREFIX}",
                "--arch=${SYSTEM_PROCESSOR}",
                "--target-os=${SYSTEM_NAME}",
                "--pkg-config=pkg-config",
                "--enable-cross-compile",
                "--disable-programs",
                "--disable-doc",
                "--enable-libx264",
                "--enable-libx265",
                "--enable-pic",
                "--enable-gpl"
            ],
            "dependencies": [
                "x264@stable",
                "x265@4.0"
            ],
            "cmake_config": "linux_shared"
        }
    ]
}
```

After building and installing, you can see the generated cmake config files as below:

```
lib
└── cmake
    └─── ffmpeg
        ├── ffmpegConfig.cmake
        ├── ffmpegConfigVersion.cmake
        ├── ffmpegModules-release.cmake
        └── ffmpegModules.cmake
```

Finally, you can use it in your cmake project as below:

```cmake
find_package(ffmpeg REQUIRED)
target_link_libraries(${PROJECT_NAME} PRIVATE
    FFmpeg::avutil
    FFmpeg::avcodec
    FFmpeg::avdevice
    FFmpeg::avfilter
    FFmpeg::avformat
    FFmpeg::postproc
    FFmpeg::swresample
    FFmpeg::swscale
)
```

> Please note that, the namespace is defined in cmake_config file, if not defined, it would be the same as the library name.