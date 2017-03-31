package v3action

import "time"

//go:generate counterfeiter . Config

type Config interface {
	PollingInterval() time.Duration
}
