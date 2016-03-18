// package metricbatcher provides a mechanism to batch counter updates into a single event.
package metricbatcher

import (
	"sync"
	"time"

	"github.com/cloudfoundry/dropsonde/metric_sender"
)

// MetricBatcher batches counter increment/add calls into periodic, aggregate events.
type MetricBatcher struct {
	metrics      map[string]uint64
	batchTicker  *time.Ticker
	metricSender metric_sender.MetricSender
	lock         sync.Mutex
	closed       bool
	closedChan   chan struct{}
}

// New instantiates a running MetricBatcher. Eventswill be emitted once per batchDuration. All
// updates to a given counter name will be combined into a single event and sent to metricSender.
func New(metricSender metric_sender.MetricSender, batchDuration time.Duration) *MetricBatcher {
	mb := &MetricBatcher{
		metrics:      make(map[string]uint64),
		batchTicker:  time.NewTicker(batchDuration),
		metricSender: metricSender,
		closed:       false,
		closedChan:   make(chan struct{}),
	}

	go func() {
		for {
			select {
			case <-mb.batchTicker.C:
				mb.flush(mb.resetAndReturnMetrics())
			case <-mb.closedChan:
				mb.batchTicker.Stop()
				return
			}
		}
	}()

	return mb
}

// BatchIncrementCounter increments the named counter by 1, but does not immediately send a
// CounterEvent.
func (mb *MetricBatcher) BatchIncrementCounter(name string) {
	mb.BatchAddCounter(name, 1)
}

// BatchAddCounter increments the named counter by the provided delta, but does not
// immediately send a CounterEvent.
func (mb *MetricBatcher) BatchAddCounter(name string, delta uint64) {
	mb.lock.Lock()
	defer mb.lock.Unlock()

	if mb.closed {
		panic("Attempting to send metrics after closed")
	}

	mb.metrics[name] += delta
}

// Reset clears the MetricBatcher's internal state, so that no counters are tracked.
func (mb *MetricBatcher) Reset() {
	mb.resetAndReturnMetrics()
}

// Closes the metrics batcher. Using the batcher after closing, will cause a panic.
func (mb *MetricBatcher) Close() {
	mb.lock.Lock()
	defer mb.lock.Unlock()

	mb.closed = true
	close(mb.closedChan)

	mb.flush(mb.unsafeResetAndReturnMetrics())
}

func (mb *MetricBatcher) flush(metrics map[string]uint64) {
	for name, delta := range metrics {
		mb.metricSender.AddToCounter(name, delta)
	}
}

func (mb *MetricBatcher) resetAndReturnMetrics() map[string]uint64 {
	mb.lock.Lock()
	defer mb.lock.Unlock()

	return mb.unsafeResetAndReturnMetrics()
}

func (mb *MetricBatcher) unsafeResetAndReturnMetrics() map[string]uint64 {
	localMetrics := mb.metrics

	mb.metrics = make(map[string]uint64, len(mb.metrics))

	return localMetrics
}
