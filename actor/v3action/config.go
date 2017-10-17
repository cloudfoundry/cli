package v3action

import "time"

//go:generate counterfeiter . Config

type Config interface {
	AccessToken() string
	PollingInterval() time.Duration
	SSHOAuthClient() string
	StartupTimeout() time.Duration
	StagingTimeout() time.Duration
}
