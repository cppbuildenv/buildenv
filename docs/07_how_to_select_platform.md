# How to select platform.

buildenv.json is the global configuration file of buildenv and `platform` is part of buildenv.json, buildenv provider two ways to select platform.

## 2. Select it by cli with arguments.

```
$ ./buildenv -select_platform x86_64-linux-20.04

[✔] ======== current platform: x86_64-linux-20.04. ========
```

## 3. Select it by cli with menus.

Execute `./buildenv` to enter cli menu mode.

```
$ ./buildenv

   Welcome to buildenv!                                   
   Please choose an option from the menu below...         
                                                          
    1. Init buildenv with conf repo.                      
    2. Create a new platform.                             
  > 3. Select your current platform.                      
    4. Create a new project.                              
    5. Select your current project.                       
    6. Create a new tool.                                 
    7. Create a new port.                                 
    8. Integrate buildenv, then you can run it everywhere.
    9. About and usage.                                   
                                                          
    ↑/k up • ↓/j down • q quit • ? more
```

Select menu 3 with up-down arrow key and press enter key:

```
$ ./buildenv

   Select your current platform:       

    1. aarch64-linux-j721e             
  > 2. x86_64-linux-20.04              
    3. x86_64-linux-native             
                                       
    ↑/k up • ↓/j down • q quit • ? more
```

Select platform to be selected and press enter key:

```
[✔] ======== current platform: x86_64-linux-20.04. ========
```