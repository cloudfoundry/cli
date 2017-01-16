// package metricbatcher provides a mechanism to batch counter updates into a single event.
package metricbatcher

import (
	"sync"
	"time"

	"github.com/cloudfoundry/dropsonde/metric_sender"
)

//go:generate hel --type MetricSender --output mock_metric_sender_test.go

type MetricSender interface {
	Counter(name string) metric_sender.CounterChainer
}

type batch struct {
	name  string
	tags  map[string]string
	value uint64
}

// MetricBatcher batches counter increment/add calls into periodic, aggregate events.
type MetricBatcher struct {
	metrics      []batch
	batchTicker  *time.Ticker
	metricSender MetricSender
	lock         sync.Mutex
	closed       bool
	closedChan   chan struct{}
}

// New instantiates a running MetricBatcher. Eventswill be emitted once per batchDuration. All
// updates to a given counter name will be combined into a single event and sent to metricSender.
func New(metricSender MetricSender, batchDuration time.Duration) *MetricBatcher {
	mb := &MetricBatcher{
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

	mb.add(batch{name: name, value: delta})
}

func (mb *MetricBatcher) add(newBatch batch) {
	for i, batch := range mb.metrics {
		if batch.name != newBatch.name {
			continue
		}
		if len(batch.tags) != len(newBatch.tags) {
			continue
		}
		tagsMatch := true
		for k, v := range batch.tags {
			if newBatch.tags[k] != v {
				tagsMatch = false
				break
			}
		}
		if !tagsMatch {
			continue
		}
		mb.metrics[i].value += newBatch.value
		return
	}
	mb.metrics = append(mb.metrics, newBatch)
}

// BatchCounter returns a BatchCounterChainer which can be used to prepare
// a counter event before batching it up.
func (mb *MetricBatcher) BatchCounter(name string) BatchCounterChainer {
	return batchCounterChainer{
		batcher: mb,
		name:    name,
		tags:    make(map[string]string),
	}
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

func (mb *MetricBatcher) flush(metrics []batch) {
	for _, metric := range metrics {
		counter := mb.metricSender.Counter(metric.name)
		for k, v := range metric.tags {
			counter.SetTag(k, v)
		}
		counter.Add(metric.value)
	}
}

func (mb *MetricBatcher) resetAndReturnMetrics() []batch {
	mb.lock.Lock()
	defer mb.lock.Unlock()

	return mb.unsafeResetAndReturnMetrics()
}

func (mb *MetricBatcher) unsafeResetAndReturnMetrics() []batch {
	localMetrics := mb.metrics
	mb.metrics = make([]batch, 0, len(mb.metrics))
	return localMetrics
}

type BatchCounterChainer interface {
	SetTag(key, value string) BatchCounterChainer
	Increment()
	Add(value uint64)
}

type batchCounterChainer struct {
	batcher *MetricBatcher
	name    string
	tags    map[string]string
}

func (c batchCounterChainer) SetTag(key, value string) BatchCounterChainer {
	c.tags[key] = value
	return c
}

func (c batchCounterChainer) Increment() {
	c.batcher.lock.Lock()
	defer c.batcher.lock.Unlock()
	c.batcher.add(batch{name: c.name, value: 1, tags: c.tags})
}

func (c batchCounterChainer) Add(value uint64) {
	c.batcher.lock.Lock()
	defer c.batcher.lock.Unlock()
	c.batcher.add(batch{name: c.name, value: value, tags: c.tags})
}
