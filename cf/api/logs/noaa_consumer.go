package logs

import (
	"github.com/cloudfoundry/sonde-go/events"
)

// Should be satisfied automatically by *noaa.Consumer
//go:generate counterfeiter . NoaaConsumer

type NoaaConsumer interface {
	TailingLogs(string, string) (<-chan *events.LogMessage, <-chan error)
	RecentLogs(appGuid string, authToken string) ([]*events.LogMessage, error)
	Close() error
	SetOnConnectCallback(cb func())
}
