package api

import (
	"github.com/cloudfoundry/noaa"
	"github.com/cloudfoundry/sonde-go/events"
)

type NoaaConsumer interface {
	GetContainerMetrics(string, string) ([]*events.ContainerMetric, error)
	RecentLogs(string, string) ([]*events.LogMessage, error)
	TailingLogs(string, string, chan<- *events.LogMessage, chan<- error)
	SetOnConnectCallback(func())
	Close() error
}

type noaaConsumer struct {
	consumer *noaa.Consumer
}

func NewNoaaConsumer(consumer *noaa.Consumer) NoaaConsumer {
	return &noaaConsumer{
		consumer: consumer,
	}
}

func (n *noaaConsumer) GetContainerMetrics(appGuid, token string) ([]*events.ContainerMetric, error) {
	return n.consumer.ContainerMetrics(appGuid, token)
}

func (n *noaaConsumer) RecentLogs(appGuid string, authToken string) ([]*events.LogMessage, error) {
	return n.consumer.RecentLogs(appGuid, authToken)
}

func (n *noaaConsumer) TailingLogs(appGuid string, authToken string, outputChan chan<- *events.LogMessage, errorChan chan<- error) {
	n.consumer.TailingLogs(appGuid, authToken, outputChan, errorChan)
}

func (n *noaaConsumer) SetOnConnectCallback(cb func()) {
	n.consumer.SetOnConnectCallback(cb)
}

func (n *noaaConsumer) Close() error {
	return n.consumer.Close()
}
