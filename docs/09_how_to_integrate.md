# How to integarte.

buildenv is just a standalone executable binary. However, to make it easier to invoke globally, we support adding the path of buildenv to the system environment variables, we can integrate buildenv via invoke cli with arguments or cli menus.

## Integrate via cli arguments.

```
./buildenv -integrate

[✔] ======== buildenv is installed.
```

## Integrate via cli menus.

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
    7. Create a new port.                                 
  > 8. Integrate buildenv, then you can run it everywhere.
    9. About and usage.                                   
                                                          
    ↑/k up • ↓/j down • q quit • ? more 
```

Select menu 8 with up-down arrow key and press enter key:

```
$ ./buildenv

Integrate buildenv.
-----------------------------------
This will append buildenv's file dir to ~/.profile, then you can use buildenv anywhere..

[↵ -> execute | ctrl+c/q -> quit]
```

```
[✔] ======== buildenv is installed.
```

Finaly, the `~/.profile` would be like below:

```
# ~/.profile: executed by the command interpreter for login shells.
# This file is not read by bash(1), if ~/.bash_profile or ~/.bash_login
# exists.
# see /usr/share/doc/bash/examples/startup-files for examples.
# the files are located in the bash-doc package.
# the default umask is set in /etc/profile; for setting the umask
# for ssh logins, install and configure the libpam-umask package.
#umask 022
# if running bash
if [ -n "$BASH_VERSION" ]; then
    # include .bashrc if it exists
    if [ -f "$HOME/.bashrc" ]; then
        . "$HOME/.bashrc"
    fi
fi
# set PATH so it includes user's private bin if it exists
if [ -d "$HOME/bin" ] ; then
    PATH="$HOME/bin:$PATH"
fi
# set PATH so it includes user's private bin if it exists
if [ -d "$HOME/.local/bin" ] ; then
    PATH="$HOME/.local/bin:$PATH"
fi

# buildenv runtime path (added by buildenv)
export PATH=/home/phil/software/buildenv:$PATH
```