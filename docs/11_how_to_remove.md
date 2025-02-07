# How to remove third-party library.

Removing third-party library in buildenv has four ways:

- **./buildenv -remove xxx**: Builenv would remove all installed files that belongs to xxx from `installed` folder, but xxx's package still exist. If you install xxx again, the installation speed would be very fast because buildenv tries to restore from the `packages` folder to the `installed` folder.

- **./buildenv -remove xxx -recursive**: Simlar to `./buildenv -remove xxx` and also will remove xxx's sub-depedencies files from `installed` folder. If you install xxx again, the installation speed would be also very fast.

- **./buildenv -purge xxx**: Simlar to `./buildenv -remove xxx` and also will remove xxx's package folder. If you install xxx again, buildenv would configure, build and install it from source.

- **./buildenv -purge xxx -recursive**: This will remove xxx's files from the `installed` folder, remove its package, and also its sub-dependencies. If you install xxx again, buildenv will configure, build, and install it from the source, along with its sub-dependencies.

>If third-paty libary has been added in project's JSON file, then you can execute `./buildenv -remove name` instead of `./buildenv -remove name@version`, for example: `./buildenv -remove x264`.