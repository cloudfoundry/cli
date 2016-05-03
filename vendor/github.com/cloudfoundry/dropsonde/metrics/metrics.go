// Package metrics provides a simple API for sending value and counter metrics
// through the dropsonde system.
//
// Use
//
// See the documentation for package dropsonde for configuration details.
//
// Importing package dropsonde and initializing will initial this package.
// To send metrics use
//
//		metrics.SendValue(name, value, unit)
//
// for sending known quantities, and
//
//		metrics.IncrementCounter(name)
//
// to increment a counter. (Note that the value of the counter is maintained by
// the receiver of the counter events, not the application that includes this
// package.)
package metrics

import (
	"github.com/cloudfoundry/dropsonde/metric_sender"
)

var metricSender metric_sender.MetricSender
var metricBatcher MetricBatcher

type MetricBatcher interface {
	BatchIncrementCounter(name string)
	BatchAddCounter(name string, delta uint64)
	Close()
}

// Initialize prepares the metrics package for use with the automatic Emitter.
func Initialize(ms metric_sender.MetricSender, mb MetricBatcher) {
	if metricBatcher != nil {
		metricBatcher.Close()
	}
	metricSender = ms
	metricBatcher = mb
}

// Closes the metrics system and flushes any batch metrics.
func Close() {
	metricBatcher.Close()
}

// SendValue sends a value event for the named metric. See
// http://metrics20.org/spec/#units for the specifications on allowed units.
func SendValue(name string, value float64, unit string) error {
	if metricSender == nil {
		return nil
	}
	return metricSender.SendValue(name, value, unit)
}

// IncrementCounter sends an event to increment the named counter by one.
// Maintaining the value of the counter is the responsibility of the receiver of
// the event, not the process that includes this package.
func IncrementCounter(name string) error {
	if metricSender == nil {
		return nil
	}
	return metricSender.IncrementCounter(name)
}

// BatchIncrementCounter increments a counter but, unlike IncrementCounter, does
// not emit a CounterEvent for each increment; instead, the increments are batched
// and a single CounterEvent is sent after the timeout.
func BatchIncrementCounter(name string) {
	if metricBatcher == nil {
		return
	}
	metricBatcher.BatchIncrementCounter(name)
}

// AddToCounter sends an event to increment the named counter by the specified
// (positive) delta. Maintaining the value of the counter is the responsibility
// of the receiver, as with IncrementCounter.
func AddToCounter(name string, delta uint64) error {
	if metricSender == nil {
		return nil
	}
	return metricSender.AddToCounter(name, delta)
}

// BatchAddCounter adds delta to a counter but, unlike AddCounter, does not emit a
// CounterEvent for each add; instead, the adds are batched and a single CounterEvent
// is sent after the timeout.
func BatchAddCounter(name string, delta uint64) {
	if metricBatcher == nil {
		return
	}
	metricBatcher.BatchAddCounter(name, delta)
}

// SendContainerMetric sends a metric that records resource usage of an app in a container.
// The container is identified by the applicationId and the instanceIndex. The resource
// metrics are CPU percentage, memory and disk usage in bytes. Returns an error if one occurs
// when sending the metric.
func SendContainerMetric(applicationId string, instanceIndex int32, cpuPercentage float64, memoryBytes uint64, diskBytes uint64) error {
	if metricSender == nil {
		return nil
	}

	return metricSender.SendContainerMetric(applicationId, instanceIndex, cpuPercentage, memoryBytes, diskBytes)
}
