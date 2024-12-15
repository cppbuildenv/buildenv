package config

import "buildenv/pkg/color"

func SyncSuccess(repoUrlInside bool) string {
	if repoUrlInside {
		return color.Sprintf(color.Magenta, "\n[✔] ======== conf repo is synchronized ========\n\n")
	} else {
		return color.Sprintf(color.Magenta, "\n[✔] ======== buildenv.json is created but need to config it later ========\n\n")
	}
}

func SyncFailed(err error) string {
	return color.Sprintf(color.Red, "\n[✘] buildenv.json is invalid.\n[☛] %s.\n\n", err)
}

func PlatformCreated(platform string) string {
	return color.Sprintf(color.Magenta, "\n[✔] ======== %s is created but need to config it later ========\n\n", platform)
}

func PlatformCreateFailed(platform string, err error) string {
	return color.Sprintf(color.Red, "\n[✘] %s could not be created.\n[☛] %s.\n\n", platform, err)
}

func PlatformSelected(platform string) string {
	return color.Sprintf(color.Magenta, "\n[✔] ======== current platform: %s ========\n\n", platform)
}

func PlatformSelectedFailed(platform string, err error) string {
	if platform == "" {
		return color.Sprintf(color.Red, "\n[✘] %s.\n\n", err)
	} else {
		return color.Sprintf(color.Red, "\n[✘] %s is broken.\n[☛] %s.\n\n", platform, err)
	}
}

func ProjectCreated(project string) string {
	return color.Sprintf(color.Magenta, "\n[✔] ======== %s is created but need to config it later ========\n\n", project)
}

func ProjectCreateFailed(project string, err error) string {
	return color.Sprintf(color.Red, "\n[✘] %s could not be created.\n[☛] %s.\n\n", project, err)
}

func ProjectSelected(project string) string {
	return color.Sprintf(color.Magenta, "\n[✔] ======== build environment is ready for project: %s ========\n\n", project)
}

func ProjectSelectedFailed(platform string, err error) string {
	if platform == "" {
		return color.Sprintf(color.Red, "\n[✘] %s.\n\n", err)
	} else {
		return color.Sprintf(color.Red, "\n[✘] %s is broken.\n[☛] %s.\n\n", platform, err)
	}
}

func IntegrateSuccessfully() string {
	return color.Sprintf(color.Magenta, "\n[✔] ======== buildenv is integrated. ========\n\n")
}

func IntegrateFailed(err error) string {
	return color.Sprintf(color.Red, "\n[✘] buildenv integrate failed.\n[☛] %s.\n\n", err)
}

func InstallSuccessfully(port string) string {
	return color.Sprintf(color.Magenta, "\n[✔] ======== %s install successfully. ========\n\n", port)
}

func InstallFailed(port string, err error) string {
	return color.Sprintf(color.Red, "\n[✘] %s install failed.\n[☛] %s.\n\n", port, err)
}

func UninstallSuccessfully(port string) string {
	return color.Sprintf(color.Magenta, "\n[✔] ======== %s uninstall successfully. ========\n\n", port)
}

func UninstallFailed(port string, err error) string {
	return color.Sprintf(color.Red, "\n[✘] %s uninstall failed.\n[☛] %s.\n\n", port, err)
}
