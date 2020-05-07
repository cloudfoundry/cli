package v3action

import (
	"time"

	"code.cloudfoundry.org/cli/util/configv3"
)

//go:generate counterfeiter . Config

type Config interface {
	AccessToken() string
	DialTimeout() time.Duration
	PollingInterval() time.Duration
	SetTargetInformation(args configv3.TargetInformationArgs)
	SetTokenInformation(accessToken string, refreshToken string, sshOAuthClient string)
	SetUAAClientCredentials(client string, clientSecret string)
	SetUAAGrantType(uaaGrantType string)
	SkipSSLValidation() bool
	SSHOAuthClient() string
	StartupTimeout() time.Duration
	StagingTimeout() time.Duration
	Target() string
	UAAGrantType() string
	UnsetOrganizationAndSpaceInformation()
}
