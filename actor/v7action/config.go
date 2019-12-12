package v7action

import "time"

//go:generate counterfeiter . Config

type Config interface {
	AccessToken() string
	DialTimeout() time.Duration
	PollingInterval() time.Duration
	RefreshToken() string
	SSHOAuthClient() string
	SetAccessToken(token string)
	SetRefreshToken(token string)
	SetTokenInformation(accessToken string, refreshToken string, token string)
	SetUAAClientCredentials(client string, clientSecret string)
	SetUAAGrantType(grantType string)
	StagingTimeout() time.Duration
	StartupTimeout() time.Duration
	UAAGrantType() string
	UnsetOrganizationAndSpaceInformation()
}
