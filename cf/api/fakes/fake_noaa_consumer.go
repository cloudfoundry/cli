package fakes

import (
	"sync"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/noaa/events"
)

type FakeNoaaConsumer struct {
	TailFunc                       func(logChan chan<- *events.LogMessage, errChan chan<- error, stopChan chan struct{})
	GetContainerMetricsStub        func(string, string) ([]*events.ContainerMetric, error)
	getContainerMetricsMutex       sync.RWMutex
	getContainerMetricsArgsForCall []struct {
		arg1 string
		arg2 string
	}
	getContainerMetricsReturns struct {
		result1 []*events.ContainerMetric
		result2 error
	}
	RecentLogsStub        func(string, string) ([]*events.LogMessage, error)
	recentLogsMutex       sync.RWMutex
	recentLogsArgsForCall []struct {
		arg1 string
		arg2 string
	}
	recentLogsReturns struct {
		result1 []*events.LogMessage
		result2 error
	}
	TailingLogsStub        func(string, string, chan<- *events.LogMessage, chan<- error, chan struct{})
	tailingLogsMutex       sync.RWMutex
	tailingLogsArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 chan<- *events.LogMessage
		arg4 chan<- error
		arg5 chan struct{}
	}
	SetOnConnectCallbackStub        func(func())
	setOnConnectCallbackMutex       sync.RWMutex
	setOnConnectCallbackArgsForCall []struct {
		arg1 func()
	}
	CloseStub        func() error
	closeMutex       sync.RWMutex
	closeArgsForCall []struct{}
	closeReturns     struct {
		result1 error
	}
}

func (fake *FakeNoaaConsumer) GetContainerMetrics(arg1 string, arg2 string) ([]*events.ContainerMetric, error) {
	fake.getContainerMetricsMutex.Lock()
	defer fake.getContainerMetricsMutex.Unlock()
	fake.getContainerMetricsArgsForCall = append(fake.getContainerMetricsArgsForCall, struct {
		arg1 string
		arg2 string
	}{arg1, arg2})
	if fake.GetContainerMetricsStub != nil {
		return fake.GetContainerMetricsStub(arg1, arg2)
	} else {
		return fake.getContainerMetricsReturns.result1, fake.getContainerMetricsReturns.result2
	}
}

func (fake *FakeNoaaConsumer) GetContainerMetricsCallCount() int {
	fake.getContainerMetricsMutex.RLock()
	defer fake.getContainerMetricsMutex.RUnlock()
	return len(fake.getContainerMetricsArgsForCall)
}

func (fake *FakeNoaaConsumer) GetContainerMetricsArgsForCall(i int) (string, string) {
	fake.getContainerMetricsMutex.RLock()
	defer fake.getContainerMetricsMutex.RUnlock()
	return fake.getContainerMetricsArgsForCall[i].arg1, fake.getContainerMetricsArgsForCall[i].arg2
}

func (fake *FakeNoaaConsumer) GetContainerMetricsReturns(result1 []*events.ContainerMetric, result2 error) {
	fake.GetContainerMetricsStub = nil
	fake.getContainerMetricsReturns = struct {
		result1 []*events.ContainerMetric
		result2 error
	}{result1, result2}
}

func (fake *FakeNoaaConsumer) RecentLogs(arg1 string, arg2 string) ([]*events.LogMessage, error) {
	fake.recentLogsMutex.Lock()
	defer fake.recentLogsMutex.Unlock()
	fake.recentLogsArgsForCall = append(fake.recentLogsArgsForCall, struct {
		arg1 string
		arg2 string
	}{arg1, arg2})
	if fake.RecentLogsStub != nil {
		return fake.RecentLogsStub(arg1, arg2)
	} else {
		return fake.recentLogsReturns.result1, fake.recentLogsReturns.result2
	}
}

func (fake *FakeNoaaConsumer) RecentLogsCallCount() int {
	fake.recentLogsMutex.RLock()
	defer fake.recentLogsMutex.RUnlock()
	return len(fake.recentLogsArgsForCall)
}

func (fake *FakeNoaaConsumer) RecentLogsArgsForCall(i int) (string, string) {
	fake.recentLogsMutex.RLock()
	defer fake.recentLogsMutex.RUnlock()
	return fake.recentLogsArgsForCall[i].arg1, fake.recentLogsArgsForCall[i].arg2
}

func (fake *FakeNoaaConsumer) RecentLogsReturns(result1 []*events.LogMessage, result2 error) {
	fake.RecentLogsStub = nil
	fake.recentLogsReturns = struct {
		result1 []*events.LogMessage
		result2 error
	}{result1, result2}
}

func (fake *FakeNoaaConsumer) TailingLogs(arg1 string, arg2 string, arg3 chan<- *events.LogMessage, arg4 chan<- error, arg5 chan struct{}) {
	fake.tailingLogsMutex.Lock()
	defer fake.tailingLogsMutex.Unlock()
	fake.tailingLogsArgsForCall = append(fake.tailingLogsArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 chan<- *events.LogMessage
		arg4 chan<- error
		arg5 chan struct{}
	}{arg1, arg2, arg3, arg4, arg5})
	if fake.TailingLogsStub != nil {
		fake.TailingLogsStub(arg1, arg2, arg3, arg4, arg5)
	}

	fake.TailFunc(arg3, arg4, arg5)
}

func (fake *FakeNoaaConsumer) TailingLogsCallCount() int {
	fake.tailingLogsMutex.RLock()
	defer fake.tailingLogsMutex.RUnlock()
	return len(fake.tailingLogsArgsForCall)
}

func (fake *FakeNoaaConsumer) TailingLogsArgsForCall(i int) (string, string, chan<- *events.LogMessage, chan<- error, chan struct{}) {
	fake.tailingLogsMutex.RLock()
	defer fake.tailingLogsMutex.RUnlock()
	return fake.tailingLogsArgsForCall[i].arg1, fake.tailingLogsArgsForCall[i].arg2, fake.tailingLogsArgsForCall[i].arg3, fake.tailingLogsArgsForCall[i].arg4, fake.tailingLogsArgsForCall[i].arg5
}

func (fake *FakeNoaaConsumer) SetOnConnectCallback(arg1 func()) {
	fake.setOnConnectCallbackMutex.Lock()
	defer fake.setOnConnectCallbackMutex.Unlock()
	fake.setOnConnectCallbackArgsForCall = append(fake.setOnConnectCallbackArgsForCall, struct {
		arg1 func()
	}{arg1})
	if fake.SetOnConnectCallbackStub != nil {
		fake.SetOnConnectCallbackStub(arg1)
	}
}

func (fake *FakeNoaaConsumer) SetOnConnectCallbackCallCount() int {
	fake.setOnConnectCallbackMutex.RLock()
	defer fake.setOnConnectCallbackMutex.RUnlock()
	return len(fake.setOnConnectCallbackArgsForCall)
}

func (fake *FakeNoaaConsumer) SetOnConnectCallbackArgsForCall(i int) func() {
	fake.setOnConnectCallbackMutex.RLock()
	defer fake.setOnConnectCallbackMutex.RUnlock()
	return fake.setOnConnectCallbackArgsForCall[i].arg1
}

func (fake *FakeNoaaConsumer) Close() error {
	fake.closeMutex.Lock()
	defer fake.closeMutex.Unlock()
	fake.closeArgsForCall = append(fake.closeArgsForCall, struct{}{})
	if fake.CloseStub != nil {
		return fake.CloseStub()
	} else {
		return fake.closeReturns.result1
	}
}

func (fake *FakeNoaaConsumer) CloseCallCount() int {
	fake.closeMutex.RLock()
	defer fake.closeMutex.RUnlock()
	return len(fake.closeArgsForCall)
}

func (fake *FakeNoaaConsumer) CloseReturns(result1 error) {
	fake.CloseStub = nil
	fake.closeReturns = struct {
		result1 error
	}{result1}
}

var _ api.NoaaConsumer = new(FakeNoaaConsumer)
