package fake

import (
	"sync"
)

type FakeMetricSender struct {
	counters         map[string]uint64
	values           map[string]Metric
	containerMetrics map[string]ContainerMetric
	sync.RWMutex
}

type Metric struct {
	Value float64
	Unit  string
}

type ContainerMetric struct {
	ApplicationId string
	InstanceIndex int32
	CpuPercentage float64
	MemoryBytes   uint64
	DiskBytes     uint64
}

func NewFakeMetricSender() *FakeMetricSender {
	return &FakeMetricSender{
		counters:         make(map[string]uint64),
		values:           make(map[string]Metric),
		containerMetrics: make(map[string]ContainerMetric),
	}
}

func (fms *FakeMetricSender) SendValue(name string, value float64, unit string) error {
	fms.Lock()
	defer fms.Unlock()
	fms.values[name] = Metric{Value: value, Unit: unit}

	return nil
}

func (fms *FakeMetricSender) IncrementCounter(name string) error {
	fms.Lock()
	defer fms.Unlock()
	fms.counters[name]++

	return nil
}

func (fms *FakeMetricSender) AddToCounter(name string, delta uint64) error {
	fms.Lock()
	defer fms.Unlock()
	fms.counters[name] = fms.counters[name] + delta

	return nil
}

func (fms *FakeMetricSender) SendContainerMetric(applicationId string, instanceIndex int32, cpuPercentage float64, memoryBytes uint64, diskBytes uint64) error {
	fms.Lock()
	defer fms.Unlock()
	fms.containerMetrics[applicationId] = ContainerMetric{ApplicationId: applicationId, InstanceIndex: instanceIndex, CpuPercentage: cpuPercentage, MemoryBytes: memoryBytes, DiskBytes: diskBytes}

	return nil
}

func (fms *FakeMetricSender) HasValue(name string) bool {
	fms.RLock()
	defer fms.RUnlock()

	_, exists := fms.values[name]
	return exists
}

func (fms *FakeMetricSender) GetValue(name string) Metric {
	fms.RLock()
	defer fms.RUnlock()

	return fms.values[name]
}

func (fms *FakeMetricSender) GetCounter(name string) uint64 {
	fms.RLock()
	defer fms.RUnlock()

	return fms.counters[name]
}

func (fms *FakeMetricSender) GetContainerMetric(applicationId string) ContainerMetric {
	fms.RLock()
	defer fms.RUnlock()

	return fms.containerMetrics[applicationId]
}

func (fms *FakeMetricSender) Reset() {
	fms.Lock()
	defer fms.Unlock()

	fms.counters = make(map[string]uint64)
	fms.values = make(map[string]Metric)
	fms.containerMetrics = make(map[string]ContainerMetric)
}
