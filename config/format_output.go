package config

import "buildenv/pkg/color"

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

func ToolCreated(project string) string {
	return color.Sprintf(color.Magenta, "\n[✔] ======== %s is created but need to config it later ========\n\n", project)
}

func ToolCreateFailed(project string, err error) string {
	return color.Sprintf(color.Red, "\n[✘] %s could not be created.\n[☛] %s.\n\n", project, err)
}

func PortCreated(project string) string {
	return color.Sprintf(color.Magenta, "\n[✔] ======== %s is created but need to config it later ========\n\n", project)
}

func PortCreateFailed(project string, err error) string {
	return color.Sprintf(color.Red, "\n[✘] %s could not be created.\n[☛] %s.\n\n", project, err)
}

func ConfigInitialized() string {
	return color.Sprintf(color.Magenta, "\n[✔] ======== init buildenv successfully ========\n\n")
}

func ConfigInitFailed(configUrl string, err error) string {
	return color.Sprintf(color.Red, "\n[✘] failed to init buildenv with %s.\n[☛] %s.\n\n", configUrl, err)
}
