package command

import (
	"time"

	"code.cloudfoundry.org/cli/util/configv3"
)

//go:generate counterfeiter . Config

// Config a way of getting basic CF configuration
type Config interface {
	APIVersion() string
	AccessToken() string
	BinaryName() string
	BinaryVersion() string
	ColorEnabled() configv3.ColorSetting
	CurrentUser() (configv3.User, error)
	DialTimeout() time.Duration
	Experimental() bool
	HasTargetedOrganization() bool
	HasTargetedSpace() bool
	Locale() string
	MinCLIVersion() string
	OverallPollingTimeout() time.Duration
	Plugins() map[string]configv3.Plugin
	PollingInterval() time.Duration
	RefreshToken() string
	SetAccessToken(token string)
	SetOrganizationInformation(guid string, name string)
	SetRefreshToken(token string)
	SetSpaceInformation(guid string, name string, allowSSH bool)
	SetTargetInformation(api string, apiVersion string, auth string, minCLIVersion string, doppler string, uaa string, routing string, skipSSLValidation bool)
	SetTokenInformation(accessToken string, refreshToken string, sshOAuthClient string)
	SkipSSLValidation() bool
	Target() string
	TargetedOrganization() configv3.Organization
	TargetedSpace() configv3.Space
	UAAOAuthClient() string
	UAAOAuthClientSecret() string
	UnsetSpaceInformation()
	UnsetOrganizationInformation()
	Verbose() (bool, []string)
}
