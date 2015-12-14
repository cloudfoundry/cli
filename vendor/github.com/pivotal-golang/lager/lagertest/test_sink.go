package lagertest

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/onsi/gomega/gbytes"

	"github.com/pivotal-golang/lager"
)

type TestLogger struct {
	lager.Logger
	*TestSink
}

type TestSink struct {
	lager.Sink
	*gbytes.Buffer
}

func NewTestLogger(component string) *TestLogger {
	logger := lager.NewLogger(component)

	testSink := NewTestSink()

	logger.RegisterSink(testSink)

	return &TestLogger{logger, testSink}
}

func NewTestSink() *TestSink {
	buffer := gbytes.NewBuffer()

	return &TestSink{
		Sink:   lager.NewWriterSink(buffer, lager.DEBUG),
		Buffer: buffer,
	}
}

func (s *TestSink) Logs() []lager.LogFormat {
	logs := []lager.LogFormat{}

	decoder := json.NewDecoder(bytes.NewBuffer(s.Buffer.Contents()))
	for {
		var log lager.LogFormat
		if err := decoder.Decode(&log); err == io.EOF {
			return logs
		} else if err != nil {
			panic(err)
		}
		logs = append(logs, log)
	}

	return logs
}
