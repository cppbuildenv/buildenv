# How to add port.

The port configuration file is stored under `workspace/conf/ports`. In this file, we descript a third-party library where to clone code from, how to build and how many other third-party libraries are dependented on by current library.

## 1. What does it loooks like.

CMake project port example: `conf/ports/glog/v0.6.0.json`:

```json
{
    "url": "ssh://git@192.168.0.2:8088/thirdpary/glog.git",
    "name": "glog",
    "version": "v0.6.0",
    "build_configs": [
        {
            "pattern": "*linux*",
            "build_tool": "cmake",
            "arguments": [
                "-DBUILD_SHARED_LIBS=ON",
                "-DBUILD_STATIC_LIBS=OFF",
                "-DBUILD_TESTING=OFF",
                "-DWITH_GTEST=OFF",
                "-DBUILD_EXAMPLES=OFF"
            ],
            "dependencies": [
                "gflags@v2.2.2"
            ]
        }
    ]
}
```

Makefile project port example:

```json
{
    "url": "ssh://git@192.168.0.2:8088/thirdpary/x264.git",
    "name": "x264",
    "version": "stable",
    "build_configs": [
        {
            "pattern": "*linux*",
            "build_tool": "make",
            "env_vars": [],
            "arguments": [
                "--host=${HOST}",
                "--sysroot=${SYSROOT}",
                "--cross-prefix=${CROSS_PREFIX}",
                "--enable-shared",
                "--disable-cli",
                "--enable-pic"
            ],
            "dependencies": [ ],
            "cmake_config": "linux_shared"
        }
    ]
}
```

**Notes**：

- **url**: In China, you may not be able to access github's repo directly, you can fork them to your own repository, so the url can be the url of your repository.
- **name**: repo's arational name.
- **version**: It can be a tag name or a branch name.
- **build_config**: Different third-party may have different kind build systems, we can define how to build them here.
    - **platform_pattern**, **project_pattern** : some third-party libraries need to turn on different configure arguments for platforms or projects. For example, project_AAA requires ffmpeg without x265 but project_BBB requires ffmpeg with x265, so we can add two extra build_config nodes with project_pattern "project_AAA" and "project_BBB".
    - **build_tool**: I would be `autotools`, `cmake`, `make`, `meson`, `ninja`. We'll support more buildsystems in the feature.
    - **env_vars**: It's optional, you can define some environments like `CXXFLAGS=-fPIC` here.
    - **arguments**: Different third-party libraries always have a lot of features need to turn on when configure them, we can define key-value to turn on or turn off them here. In fact, buildenv always add a lot of extra key-values for every buildsystem, like `CMAKE_PREFIX_PATH`, `CMAKE_INSTALL_PREFIX` for cmake prject and `--prefix` for makefile project. Because the parameters required for cross-compiling Makefile projects are often less standardized than those in CMake, we have predefined common dynamic variable placeholders in BuildEnv to facilitate flexible configuration, they are `${HOST}`, `${SYSTEM_NAME}`, `${SYSTEM_PROCESSOR}`, `${SYSROOT}`, `${CROSS_PREFIX}`, in fact, their value come from `toolchain` that defined in platform JSON file.
    - **dependencies**: If your third-party library has depedencies on other third-party librarys, you need to define them here, then the depedencies would be clone, configure, build and install in front of current library. Be carefull, the dependency format is `name@version`, we must exactly specify which version should be used by current library.
    - **cmake_config**: Not all third-party libraries can build by CMake. For those libraries CMake may provider FindXXX.cmake, they may not always work and sometimes require custom modifications, even some are not provided at all. The good news is buildenv can generate cmake config files for those libraries.

## 2. Create it by cli with arguments.

```
./buildenv -create_port glog@v0.6.0

[✔] ======== glog@v0.6.0 is created but need to config it later. ========
```

## 3. Create it by cli with menus.

Execute `./buildenv` to enter cli menu mode.

```
$ ./buildenv

   Welcome to buildenv!                                   
   Please choose an option from the menu below...         
                                                          
    1. Init buildenv with conf repo.                      
    2. Create a new platform.                             
    3. Select your current platform.                      
    4. Create a new project.                              
    5. Select your current project.                       
    6. Create a new tool.                                 
  > 7. Create a new port.                                 
    8. Integrate buildenv, then you can run it everywhere.
    9. About and usage.                                   
                                                                                          
    ↑/k up • ↓/j down • q quit • ? more                   
```

Select menu 7 with up-down arrow key and press enter key:

```
 Please enter your port's name: 

> your port's name        

[enter -> execute | esc -> back | ctrl+c/q -> quit]
```

Enter your new port's name and press enter key:

```
[✔] ======== glog@v0.6.0 is created but need to config it later. ========
```