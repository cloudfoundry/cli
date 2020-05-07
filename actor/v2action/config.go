package v2action

import (
	"time"

	"code.cloudfoundry.org/cli/util/configv3"
)

//go:generate counterfeiter . Config

type Config interface {
	AccessToken() string
	BinaryName() string
	DialTimeout() time.Duration
	PollingInterval() time.Duration
	RefreshToken() string
	SetAccessToken(accessToken string)
	SetRefreshToken(refreshToken string)
	SetTargetInformation(args configv3.TargetInformationArgs)
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
