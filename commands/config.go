package commands

import "code.cloudfoundry.org/cli/utils/config"

//go:generate counterfeiter . Config

// Config a way of getting basic CF configuration
type Config interface {
	BinaryName() string
	ColorEnabled() config.ColorSetting
	Locale() string
	Plugins() map[string]config.Plugin
	SetTargetInformation(api string, apiVersion string, auth string, loggregator string, doppler string, uaa string)
}
