package v7action

import "time"

//go:generate counterfeiter . Config

type Config interface {
	AccessToken() string
	RefreshToken() string
	DialTimeout() time.Duration
	PollingInterval() time.Duration
	SSHOAuthClient() string
	StartupTimeout() time.Duration
	StagingTimeout() time.Duration
}
