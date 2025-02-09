# How to select project.

buildenv.json is the global configuration file of buildenv and `project` is part of buildenv.json, buildenv provider two ways to select project.

## 2. Select it by cli with arguments.

```
$ ./buildenv select --project=project_001

[✔] ======== current project: project_001. ========
```

## 3. Select it by cli with menus.

Execute `./buildenv` to enter cli menu mode.

```
$ ./buildenv

   Welcome to buildenv!                                   
   Please choose an option from the menu below...         
                                                          
    1. Init buildenv with conf repo.                      
    2. Create a new platform.                             
    3. Select your current platform.                      
    4. Create a new project.                              
  > 5. Select your current project.                       
    6. Create a new tool.                                 
    7. Create a new port.                                 
    8. Integrate buildenv, then you can run it everywhere.
    9. About and usage.                                   
                                                          
    ↑/k up • ↓/j down • q quit • ? more
```

Select menu 5 with up-down arrow key and press enter key:

```
$ ./buildenv

   Select your current project:        
                                       
  > 1. project_001                     
    2. project_002                     
    3. project_003                     
                                       
    ↑/k up • ↓/j down • q quit • ? more
```

Select project to be selected and press enter key:

```
[✔] ======== buildenv is ready for project: project_001. ========
```