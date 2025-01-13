# How to add new tool.

The tool configuration file is stored under `workspace/conf/tools`. The purpose of the tool configuration is to download the tool, extract it, and then add its bin file path to the PATH, so that they can directly invoke it during the compilation process.

## 1. What does it loooks like.

for example: `conf/tools/nasm-2.16.03.json`:

```json
{
    "url": "http://192.168.0.1:8080/build_resource/nasm-2.16.03.tar.gz",
    "archive_name":"nasm-2.16.03-x86_64-linux.tar.gz",
    "path": "nasm-2.16.03-x86_64-linux/bin"
}
```

**Notes**:

- url: It can be a url of http, https or ftp, BuildEnv will download it. It also can be a local file path, and should has a prefix "file:///", for example: `file:////home/phil/buildresource/nasm-2.16.03/bin`.
- archive_name: you can change archive's original file name.
- path: It is typically extracted from a compressed file to an internal path, usually pointing to the directory where the internal bin is located.

## 2. Create it by cli with arguments.

```
./buildenv -create_tool cmake-3.30.5-linux-x86_64

[✔] ========  cmake-3.30.5-linux-x86_64 is created but need to config it later. ========
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
  > 6. Create a new tool.                                 
    7. Create a new port.                                 
    8. Integrate buildenv, then you can run it everywhere.
    9. About and usage.                                   
                                                          
    ↑/k up • ↓/j down • q quit • ? more       
```

Select menu 6 with up-down arrow key and press enter key:

```
$ ./buildenv

 Please enter your tool's name:                   

> your tool's name...                         
[enter -> execute | esc -> back | ctrl+c/q -> quit]
```

Enter your new tool's name and press enter key:

```
[✔] ========  cmake-3.30.5-linux-x86_64 is created but need to config it later. ========
```