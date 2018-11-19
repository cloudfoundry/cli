package v2action

import "time"

//go:generate counterfeiter . Config

type Config interface {
	AccessToken() string
	BinaryName() string
	DialTimeout() time.Duration
	PollingInterval() time.Duration
	RefreshToken() string
	SetAccessToken(accessToken string)
	SetRefreshToken(refreshToken string)
	SetTargetInformation(api string, apiVersion string, auth string, minCLIVersion string, doppler string, routing string, skipSSLValidation bool)
	SetTokenInformation(accessToken string, refreshToken string, sshOAuthClient string)
	SetUAAClientCredentials(client string, clientSecret string)
	SetUAAGrantType(uaaGrantType string)
	SkipSSLValidation() bool
	SSHOAuthClient() string
	StagingTimeout() time.Duration
	StartupTimeout() time.Duration
	Target() string
	UAAGrantType() string
	UnsetOrganizationAndSpaceInformation()
	UnsetSpaceInformation()
	Verbose() (bool, []string)
}
