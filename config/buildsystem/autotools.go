package buildsystem

func NewAutoTool(config BuildConfig) *cmake {
	return &cmake{BuildConfig: config}
}

type autoTool struct {
	BuildConfig
}

func (a autoTool) Configure() error {
	return nil
}

func (a autoTool) Build() error {
	return nil
}

func (a autoTool) Install() error {
	return nil
}
