# How to add new project.

The tool configuration file is stored under `workspace/conf/projects`. In this file, we define third-party libraries that required by current project. Also we can define CMake global key-values, even environemnt key-values and C/C++ micro key-values.

# What does it looks like.

for example: `conf/projects/project_001.json`:

```json
{
    "ports": [
        "x264@stable",
        "sqlite3@v3.49.0",
        "x265@4.0",
        "ffmpeg@3.4.13",
        "zlib@v1.3.1",
        "opencv@3.4.18"
    ],
    "cmake_vars": [
        "CMAKE_VAR1=value1",
        "CMAKE_VAR2=value2"
    ],
    "env_vars": [
        "ENV_VAR1=/home/ubuntu/ccache"
    ],
    "micro_vars": [
        "MICRO_VAR1=111",
        "MICRO_VAR2"
    ]
}
```

**Notes**:

- **ports**: In FFmpeg’s port file, if FFmpeg has defined dependencies on x264 and x265, defining x264 and x265 here is not mandatory.

## 2. Create it by cli with arguments.

```
$ ./buildenv create --project=project_003

[✔] ======== project_003 is created but need to config it later. ========
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
  > 4. Create a new project.                              
    5. Select your current project.                       
    6. Create a new tool.                                 
    7. Create a new port.                                 
    8. Integrate buildenv, then you can run it everywhere.
    9. About and usage.                                   
                                                          
                                                          
    ↑/k up • ↓/j down • q quit • ? more     
```

Select menu 4 with up-down arrow key and press enter key:

```
$ ./buildenv

 Please enter your project's name:                

> your project's name...

[enter -> execute | esc -> back | ctrl+c/q -> quit]
```

Enter your new platrorm's name and enter:

```
[✔] ======== project_003 is created but need to config it later. ========
```