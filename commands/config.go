package commands

import "code.cloudfoundry.org/cli/utils/config"

//go:generate counterfeiter . Config

// Config a way of getting basic CF configuration
type Config interface {
	ColorEnabled() config.ColorSetting
	PluginConfig() map[string]config.PluginConfig
}
