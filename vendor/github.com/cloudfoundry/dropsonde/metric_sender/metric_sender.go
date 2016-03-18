package metric_sender

import (
	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/sonde-go/events"
)

// A MetricSender emits metric events.
type MetricSender interface {
	SendValue(name string, value float64, unit string) error
	IncrementCounter(name string) error
	AddToCounter(name string, delta uint64) error
	SendContainerMetric(applicationId string, instanceIndex int32, cpuPercentage float64, memoryBytes uint64, diskBytes uint64) error
}

type metricSender struct {
	eventEmitter emitter.EventEmitter
}

// NewMetricSender instantiates a metricSender with the given EventEmitter.
func NewMetricSender(eventEmitter emitter.EventEmitter) MetricSender {
	return &metricSender{eventEmitter: eventEmitter}
}

// SendValue sends a metric with the given name, value and unit. See
// http://metrics20.org/spec/#units for a specification of acceptable units.
// Returns an error if one occurs while sending the event.
func (ms *metricSender) SendValue(name string, value float64, unit string) error {
	return ms.eventEmitter.Emit(&events.ValueMetric{Name: &name, Value: &value, Unit: &unit})
}

// IncrementCounter sends an event to increment the named counter by one.
// Maintaining the value of the counter is the responsibility of the receiver of
// the event, not the process that includes this package.
func (ms *metricSender) IncrementCounter(name string) error {
	return ms.AddToCounter(name, 1)
}

// AddToCounter sends an event to increment the named counter by the specified
// (positive) delta. Maintaining the value of the counter is the responsibility
// of the receiver, as with IncrementCounter.
func (ms *metricSender) AddToCounter(name string, delta uint64) error {
	return ms.eventEmitter.Emit(&events.CounterEvent{Name: &name, Delta: &delta})
}

// SendContainerMetric sends a metric that records resource usage of an app in a container.
// The container is identified by the applicationId and the instanceIndex. The resource
// metrics are CPU percentage, memory and disk usage in bytes. Returns an error if one occurs
// when sending the metric.
func (ms *metricSender) SendContainerMetric(applicationId string, instanceIndex int32, cpuPercentage float64, memoryBytes uint64, diskBytes uint64) error {
	return ms.eventEmitter.Emit(&events.ContainerMetric{ApplicationId: &applicationId, InstanceIndex: &instanceIndex, CpuPercentage: &cpuPercentage, MemoryBytes: &memoryBytes, DiskBytes: &diskBytes})
}
