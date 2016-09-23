package commands

import "code.cloudfoundry.org/cli/utils/config"

//go:generate counterfeiter . Config

// Config a way of getting basic CF configuration
type Config interface {
	APIVersion() string
	BinaryName() string
	ColorEnabled() config.ColorSetting
	CurrentUser() (config.User, error)
	Locale() string
	Plugins() map[string]config.Plugin
	SetTargetInformation(api string, apiVersion string, auth string, loggregator string, doppler string, uaa string, routing string, skipSSLValidation bool)
	SetTokenInformation(accessToken string, refreshToken string, sshOAuthClient string)
	Target() string
	TargetedOrganization() config.Organization
	TargetedSpace() config.Space
	Experimental() bool
}
