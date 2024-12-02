package config

import "buildenv/pkg/color"

func SyncSuccess(repoUrlInside bool) string {
	if repoUrlInside {
		return color.Sprintf(color.Blue, "[✔] ======== conf repo is synchronized.\n")
	} else {
		return color.Sprintf(color.Blue, "[✔] ======== buildenv.json is created but need to config it later.\n")
	}
}

func SyncFailed(err error) string {
	return color.Sprintf(color.Red, "[✘] buildenv.json is invalid.\n[?] %s.\n", err)
}

func PlatformCreated(platform string) string {
	return color.Sprintf(color.Blue, "[✔] ======== %s is created but need to config it later.\n", platform)
}

func PlatformCreateFailed(platform string, err error) string {
	return color.Sprintf(color.Red, "[✘] %s could not be created.\n[?] %s.\n", platform, err)
}

func PlatformSelected(platform string) string {
	return color.Sprintf(color.Blue, "[✔] ======== %s is prepared as your buildenv.\n", platform)
}

func PlatformSelectedFailed(platform string, err error) string {
	if platform == "" {
		return color.Sprintf(color.Red, "[✘] %s.\n", err)
	} else {
		return color.Sprintf(color.Red, "[✘] %s is broken.\n[?] %s.\n", platform, err)
	}
}

func IntegrateSuccess() string {
	return color.Sprintf(color.Blue, "[✔] ======== buildenv is integrated.\n")
}

func IntegrateFailed(err error) string {
	return color.Sprintf(color.Red, "[✘] buildenv integrate failed.\n[?] %s.\n", err)
}

func InstallFailed(port string, err error) string {
	return color.Sprintf(color.Red, "[✘] %s install failed.\n[?] %s.\n", port, err)
}
