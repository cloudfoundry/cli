package v2action

import "time"

//go:generate counterfeiter . Config

type Config interface {
	OverallPollingTimeout() time.Duration
	PollingInterval() time.Duration
	SetTargetInformation(api string, apiVersion string, auth string, loggregator string, minCLIVersion string, doppler string, uaa string, routing string, skipSSLValidation bool)
	SetTokenInformation(accessToken string, refreshToken string, sshOAuthClient string)
	SkipSSLValidation() bool
	Target() string
}
