# How to uninstall third-party library.

Uninstalling third-party library in buildenv has four ways:

- **./buildenv -uninstall xxx**: Builenv would remove all installed files that belongs to xxx from `installed` folder, but xxx's package still exist. If you install xxx again, the installation speed would be very fast because BuildEnv tries to restore from the `packages` folder to the `installed` folder.

- **./buildenv -uninstall xxx -recursive**: Simlar to `./buildenv -uninstall xxx` and also will remove xxx's sub-depedencies files from `installed` folder. If you install xxx again, the installation speed would be also very fast.

- **./buildenv -uninstall xxx -purge**: Simlar to `./buildenv -uninstall xxx` and also will remove xxx's package folder. If you install xxx again, BuildEnv would configure, build and install it from begining.

- **./buildenv -uninstall xxx -purge -recursive**: This will remove xxx's files from the `installed` folder, remove its package, and also its sub-dependencies. If you install xxx again, BuildEnv will configure, build, and install it from the beginning, along with its sub-dependencies.

>If third-paty libary has been added in project's JSON file, then you can execute `./buildenv -uninstall name` instead of `./buildenv -uninstall name@version`, for example: `./buildenv -uninstall x264`.