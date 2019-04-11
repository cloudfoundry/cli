package v2action

import "github.com/cloudfoundry/sonde-go/events"

//go:generate counterfeiter . NOAAClient

// NOAAClient is a client for getting logs.
type NOAAClient interface {
	Close() error
	RecentLogs(appGUID string, authToken string) ([]*events.LogMessage, error)
	SetOnConnectCallback(cb func())
	TailingLogs(appGUID, authToken string) (<-chan *events.LogMessage, <-chan error)
}
