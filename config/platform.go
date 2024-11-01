package config

type PlatformCallbacks interface {
	OnCreatePlatform(platformName string) error
	OnSelectPlatform(filePath string) error
}
