package v2action

import "time"

//go:generate counterfeiter . Config

type Config interface {
	AccessToken() string
	PollingInterval() time.Duration
	RefreshToken() string
	SSHOAuthClient() string
	SetAccessToken(accessToken string)
	SetRefreshToken(refreshToken string)
	SetTargetInformation(api string, apiVersion string, auth string, minCLIVersion string, doppler string, routing string, skipSSLValidation bool)
	SetTokenInformation(accessToken string, refreshToken string, sshOAuthClient string)
	SkipSSLValidation() bool
	StagingTimeout() time.Duration
	StartupTimeout() time.Duration
	Target() string
	UnsetOrganizationInformation()
	UnsetSpaceInformation()
	Verbose() (bool, []string)
}
