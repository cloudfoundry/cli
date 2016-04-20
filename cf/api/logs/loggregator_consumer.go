package logs

import (
	"github.com/cloudfoundry/loggregator_consumer"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
)

//go:generate counterfeiter . LoggregatorConsumer

type LoggregatorConsumer interface {
	Tail(appGuid string, authToken string) (<-chan *logmessage.LogMessage, error)
	Recent(appGuid string, authToken string) ([]*logmessage.LogMessage, error)
	Close() error
	SetOnConnectCallback(func())
	SetDebugPrinter(loggregator_consumer.DebugPrinter)
}
