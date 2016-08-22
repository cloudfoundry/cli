package commands

import "code.cloudfoundry.org/cli/utils/config"

//go:generate counterfeiter . Config

// Config a way of getting basic CF configuration
type Config interface {
	BinaryName() string
	ColorEnabled() config.ColorSetting
	Locale() string
	PluginConfig() map[string]config.PluginConfig
}
