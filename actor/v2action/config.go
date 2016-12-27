package v2action

import "time"

//go:generate counterfeiter . Config

type Config interface {
	PollingInterval() time.Duration
	OverallPollingTimeout() time.Duration
}
