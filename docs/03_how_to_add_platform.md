# How to add new platform.

The platform configuration file is stored under `workspace/conf/platforms`. This file defines the toolchain, rootfs, tools, and ports required for this platform.

## 1. What does it loooks like.

for example: `conf/platforms/x86_64-linux-20.04.json`:

```json
{
    "rootfs": {
        "url": "http://192.168.0.1:8080/build_resource/ubuntu-base-20.04.5/ubuntu-base-20.04.5-base-amd64.tar.gz",
        "path": "ubuntu-base-20.04.5-base-amd64",
        "pkg_config_libdir": [
            "usr/lib/x86_64-linux-gnu/pkgconfig",
            "usr/share/pkgconfig",
            "usr/lib/pkgconfig"
        ]
    },
    "toolchain": {
        "url": "http://192.168.0.1:8080/build_resource/ubuntu-base-20.04.5/gcc-9.5.0.tar.gz",
        "path": "gcc-9.5.0/bin",
        "system_name": "Linux",
        "system_processor": "x86_64",
        "host": "x86_64-linux-gnu",
        "toolchain_prefix": "x86_64-linux-gnu-",
        "cc": "x86_64-linux-gnu-gcc",
        "cxx": "x86_64-linux-gnu-g++",
        "fc": "x86_64-linux-gnu-gfortran",
        "ranlib": "x86_64-linux-gnu-ranlib",
        "ar": "x86_64-linux-gnu-ar",
        "nm": "x86_64-linux-gnu-nm",
        "objdump": "x86_64-linux-gnu-objdump",
        "strip": "x86_64-linux-gnu-strip"
    },
    "tools": [
        "cmake-3.30.5-linux-x86_64",
        "nasm-2.16.03"
    ]
}
```

**Notes:**

- url: It can be a url of http, https or ftp, buildenv will download it. It also can be a local file path, and should has a prefix "file:///", for example: `file:////home/phil/buildresource/ubuntu-base-20.04.5/gcc-9.5.0`.
- path: It is typically extracted from a compressed file to an internal path, usually pointing to the directory where the internal bin is located.

## 2. Create it by cli with arguments.

```
$ ./buildenv create --platform=x86_64-linux-22.04

[✔] ======== x86_64-linux-22.04 is created but need to config it later. ========
```

## 3. Create it by cli with menus.

Execute `./buildenv` to enter cli menu mode.

```
$ ./buildenv

   Welcome to buildenv!                                   
   Please choose an option from the menu below...         
                                                          
    1. Init buildenv with conf repo.                      
  > 2. Create a new platform.                             
    3. Select your current platform.                      
    4. Create a new project.                              
    5. Select your current project.                       
    6. Create a new tool.                                 
    7. Create a new port.                                 
    8. Integrate buildenv, then you can run it everywhere.
    9. About and usage.                                   
                                                          
                                                          
    ↑/k up • ↓/j down • q quit • ? more         
```

Select menu 2 with up-down arrow key and press enter key:

```
$ ./buildenv

 Please enter your platform's name:               

> for example: x86_64-linux-ubuntu-20.04...         

[enter -> execute | esc -> back | ctrl+c/q -> quit]
```

Enter your new platrorm's name and enter:

```
[✔] ======== x86_64-linux-22.04 is created but need to config it later. ========
```