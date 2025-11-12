package v7action

import (
	"time"

	"code.cloudfoundry.org/cli/v8/util/configv3"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Config

type Config interface {
	AccessToken() string
	APIVersion() string
	CurrentUser() (configv3.User, error)
	DialTimeout() time.Duration
	PollingInterval() time.Duration
	RefreshToken() string
	SSHOAuthClient() string
	SetAccessToken(token string)
	SetRefreshToken(token string)
	SetTargetInformation(args configv3.TargetInformationArgs)
	SetTokenInformation(accessToken string, refreshToken string, token string)
	SetUAAClientCredentials(client string, clientSecret string)
	SetUAAGrantType(grantType string)
	SkipSSLValidation() bool
	StagingTimeout() time.Duration
	StartupTimeout() time.Duration
	Target() string
	UAAGrantType() string
	UnsetOrganizationAndSpaceInformation()
	SetKubernetesAuthInfo(authInfo string)
	IsCFOnK8s() bool
}
