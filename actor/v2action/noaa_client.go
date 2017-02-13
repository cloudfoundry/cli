package v2action

import "github.com/cloudfoundry/sonde-go/events"

//go:generate counterfeiter . NOAAClient

// NOAAClient is a client for getting logs.
type NOAAClient interface {
	Close() error
	TailingLogs(appGuid, authToken string) (<-chan *events.LogMessage, <-chan error)
}
