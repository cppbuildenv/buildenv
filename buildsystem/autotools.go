package buildsystem

func NewAutoTool(config BuildConfig) *autoTool {
	return &autoTool{BuildConfig: config}
}

type autoTool struct {
	BuildConfig
}

func (a autoTool) Configure(buildType string) error {
	// Replace placeholders with real paths and values.
	a.replaceHolders()
	return nil
}

func (a autoTool) Build() error {
	return nil
}

func (a autoTool) Install() error {
	return nil
}
