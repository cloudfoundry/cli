package api

import (
	"github.com/cloudfoundry/noaa"
	"github.com/cloudfoundry/noaa/events"
)

type NoaaConsumer interface {
	GetContainerMetrics(string, string) ([]*events.ContainerMetric, error)
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
