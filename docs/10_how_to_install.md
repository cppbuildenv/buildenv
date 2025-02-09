# How to install third-party library.

**./buildenv install name@version**: Buildenv would clone library's code, then configure, build and install it. If current library has sub-dependeicies, the sub-depedencies would be cloned, configured, built and installed if front of current libary.
Finally all third-party would be installed into `installed` folder, and every third-party's also have a individual package in `packages` folder.

>If third-paty libary has been added in project's JSON file, then you can execute `./buildenv -install name` instead of `./buildenv install name@version`, for example: `./buildenv install x264`.