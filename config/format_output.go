package config

import "buildenv/pkg/color"

func SyncSuccess(repoUrlInside bool) string {
	if repoUrlInside {
		return color.Sprintf(color.Magenta, "[✔] ======== conf repo is synchronized ========\n")
	} else {
		return color.Sprintf(color.Magenta, "[✔] ======== buildenv.json is created but need to config it later ========\n")
	}
}

func SyncFailed(err error) string {
	return color.Sprintf(color.Red, "[✘] buildenv.json is invalid.\n[☛] %s.\n", err)
}

func PlatformCreated(platform string) string {
	return color.Sprintf(color.Magenta, "[✔] ======== %s is created but need to config it later ========\n", platform)
}

func PlatformCreateFailed(platform string, err error) string {
	return color.Sprintf(color.Red, "[✘] %s could not be created.\n[☛] %s.\n", platform, err)
}

func PlatformSelected(platform string) string {
	return color.Sprintf(color.Magenta, "[✔] ======== current platform: %s ========\n", platform)
}

func PlatformSelectedFailed(platform string, err error) string {
	if platform == "" {
		return color.Sprintf(color.Red, "[✘] %s.\n", err)
	} else {
		return color.Sprintf(color.Red, "[✘] %s is broken.\n[☛] %s.\n", platform, err)
	}
}

func ProjectCreated(project string) string {
	return color.Sprintf(color.Magenta, "[✔] ======== %s is created but need to config it later ========", project)
}

func ProjectCreateFailed(project string, err error) string {
	return color.Sprintf(color.Red, "[✘] %s could not be created.\n[☛] %s.\n", project, err)
}

func ProjectSelected(project string) string {
	return color.Sprintf(color.Magenta, "[✔] ======== build environment is ready for project: %s ========\n", project)
}

func ProjectSelectedFailed(platform string, err error) string {
	if platform == "" {
		return color.Sprintf(color.Red, "[✘] %s.\n", err)
	} else {
		return color.Sprintf(color.Red, "[✘] %s is broken.\n[☛] %s.\n", platform, err)
	}
}

func IntegrateSuccess() string {
	return color.Sprintf(color.Magenta, "[✔] ======== buildenv is integrated ========\n")
}

func IntegrateFailed(err error) string {
	return color.Sprintf(color.Red, "[✘] buildenv integrate failed.\n[☛] %s.\n", err)
}

func InstallFailed(port string, err error) string {
	return color.Sprintf(color.Red, "[✘] %s install failed.\n[☛] %s.\n", port, err)
}
