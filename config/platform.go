package config

type PlatformCallbacks interface {
	OnCreatePlatform(platformName string) error
	OnSelectPlatform(platformName string) error
}
