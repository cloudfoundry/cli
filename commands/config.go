package commands

import "code.cloudfoundry.org/cli/utils/configv3"

//go:generate counterfeiter . Config

// Config a way of getting basic CF configuration
type Config interface {
	APIVersion() string
	BinaryName() string
	ColorEnabled() configv3.ColorSetting
	CurrentUser() (configv3.User, error)
	Locale() string
	Plugins() map[string]configv3.Plugin
	SetTargetInformation(api string, apiVersion string, auth string, loggregator string, doppler string, uaa string, routing string, skipSSLValidation bool)
	SetTokenInformation(accessToken string, refreshToken string, sshOAuthClient string)
	Target() string
	TargetedOrganization() configv3.Organization
	TargetedSpace() configv3.Space
	Experimental() bool
}
