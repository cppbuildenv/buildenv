package console

import "buildenv/pkg/color"

func PlatformCreated(platform string) string {
	return color.Sprintf(color.Blue, "[✔] ======== %s is created but still need to config it later.\n", platform)
}

func PlatformCreateFailed(platform string, err error) string {
	return color.Sprintf(color.Red, "[✘] ======== %s could not be created: %s.\n", platform, err)
}

func PlatformSelected(platform string) string {
	return color.Sprintf(color.Blue, "[✔] ======== %s is selected as build platform.\n", platform)
}

func PlatformSelectedFailed(platform string, err error) string {
	return color.Sprintf(color.Red, "[✘] ======== %s is invalid: %s.\n", platform, err)
}

func InstallSuccess() string {
	return color.Sprintf(color.Blue, "[✔] ======== buildenv is installed.\n")
}

func InstallFailed(err error) string {
	return color.Sprintf(color.Red, "[✘] ======== buildenv install failed: %s.\n", err)
}
