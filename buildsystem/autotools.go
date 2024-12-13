package buildsystem

func NewAutoTool(config BuildConfig) *autoTool {
	return &autoTool{BuildConfig: config}
}

type autoTool struct {
	BuildConfig
}

func (a autoTool) Configure(buildType string) (string, error) {
	return "", nil
}

func (a autoTool) Build() (string, error) {
	return "", nil
}

func (a autoTool) Install() (string, error) {
	return "", nil
}

func (a autoTool) InstalledFiles(installLogFile string) ([]string, error) {
	return nil, nil
}
