# How to init buildenv.

Buildenv depends on a set of configurations, which describe the toolchain, rootfs, cmake, tools used, and the third-party libraries it depends on. Then, buildenv will download resources, pull code, compile and build tools, and install them to the specified directory based on the configuration. This file is called buildenv.json.  

## 1. What does buildens.json look like.

The generated `buildenv.json` would be like below:

```json
{
    "conf_repo_url": "ssh://git@192.168.12.34:8890/buildenv-conf.git",
    "conf_repo_ref": "develop",
    "platform_name": "",
    "project_name": "",
    "job_num": 32,
    "cache_dirs": []
}
```
> `platform_name`, `project_name` and `cache_dirs` are empyt, this requires other configurations later, please refer [05_how_to_select_platform](./05_how_to_select_platform.md) and [07_how_to_select_project](./07_how_to_select_project.md).

## 2. Init by cli argments.

```
$ ./buildenv -init -conf_repo_url=ssh://git@192.168.12.34:8890/buildenvconf.git -conf_repo_ref=master
HEAD is now at 5a024af update config
Already on 'develop'
Your branch is up to date with 'origin/develop'.
Already up to date.

[✔] ======== init buildenv successfully. ========
```

## 3. Init by cli menus.

```
$ ./buildenv

   Welcome to buildenv!                                   
   Please choose an option from the menu below...         
                                                          
  > 1. Init buildenv with conf repo.                      
    2. Create a new platform.                             
    3. Select your current platform.                      
    4. Create a new project.                              
    5. Select your current project.                       
    6. Create a new tool.                                 
    7. Create a new port.                                 
    8. Integrate buildenv, then you can run it everywhere.
    9. About and usage.                                   
                                                          
    ↑/k up • ↓/j down • q quit • ? more   
```

Then select the first menu and press enter key:

```
Initializing buildenv:

Config repo url               
ssh://git@192.168.12.34:8890/buildenv-conf.git                                                            

Config repo ref               
master               

[enter -> execute | esc -> back | ctrl+c/q -> quit]
```