package logs

import (
	"github.com/cloudfoundry/sonde-go/events"
)

// Should be satisfied automatically by *noaa.Consumer
//go:generate counterfeiter -o fakes/fake_noaa_consumer.go . NoaaConsumer
type NoaaConsumer interface {
	TailingLogs(appGuid string, authToken string, outputChan chan<- *events.LogMessage, errorChan chan<- error)
	RecentLogs(appGuid string, authToken string) ([]*events.LogMessage, error)
	Close() error
	SetOnConnectCallback(cb func())
}
