package command

import "buildenv/pkg/color"

func SyncSuccess(repoUrlInside bool) string {
	if repoUrlInside {
		return color.Sprintf(color.Blue, "[✔] ======== conf repo is synchronized.\n")
	} else {
		return color.Sprintf(color.Blue, "[✔] ======== buildenv.json is created but need to config it later.\n")
	}
}

func SyncFailed(err error) string {
	return color.Sprintf(color.Red, "[✘] ======== buildenv.json is invalid: %s.\n", err)
}

func PlatformCreated(platform string) string {
	return color.Sprintf(color.Blue, "[✔] ======== %s is created but need to config it later.\n", platform)
}

func PlatformCreateFailed(platform string, err error) string {
	return color.Sprintf(color.Red, "[✘] ======== %s could not be created: %s.\n", platform, err)
}

func PlatformSelected(platform string) string {
	return color.Sprintf(color.Blue, "[✔] ======== %s is selected as build platform.\n", platform)
}

func PlatformSelectedFailed(platform string, err error) string {
	if platform == "" {
		return color.Sprintf(color.Red, "[✘] ======== %s.\n", err)
	} else {
		return color.Sprintf(color.Red, "[✘] ======== %s is invalid: %s.\n", platform, err)
	}
}

func InstallSuccess() string {
	return color.Sprintf(color.Blue, "[✔] ======== buildenv is installed.\n")
}

func InstallFailed(err error) string {
	return color.Sprintf(color.Red, "[✘] ======== buildenv install failed: %s.\n", err)
}
