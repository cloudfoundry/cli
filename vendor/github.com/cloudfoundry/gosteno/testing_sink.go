package gosteno

import (
	"sync"
)

type TestingSink struct {
	records []*Record

	sync.RWMutex
}

var theGlobalTestSink *TestingSink
var globalSyncMutex = &sync.RWMutex{}

func EnterTestMode(logLevel ...LogLevel) {
	globalSyncMutex.Lock()
	defer globalSyncMutex.Unlock()

	theGlobalTestSink = NewTestingSink()

	stenoConfig := Config{
		Sinks: []Sink{theGlobalTestSink},
	}

	if len(logLevel) > 0 {
		stenoConfig.Level = logLevel[0]
	}

	Init(&stenoConfig)
}

func GetMeTheGlobalTestSink() *TestingSink {
	globalSyncMutex.RLock()
	defer globalSyncMutex.RUnlock()

	return theGlobalTestSink
}

func NewTestingSink() *TestingSink {
	return &TestingSink{
		records: make([]*Record, 0, 10),
	}
}

func (tSink *TestingSink) AddRecord(record *Record) {
	tSink.Lock()
	defer tSink.Unlock()

	tSink.records = append(tSink.records, record)
}

func (tSink *TestingSink) Flush() {

}

func (tSink *TestingSink) SetCodec(codec Codec) {

}

func (tSink *TestingSink) GetCodec() Codec {
	return nil
}

func (tSink *TestingSink) Records() []*Record {
	tSink.RLock()
	defer tSink.RUnlock()

	return tSink.records
}
