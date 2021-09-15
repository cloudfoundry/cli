package lager

import "code.cloudfoundry.org/lager/internal/truncate"

type truncatingSink struct {
	sink                Sink
	maxDataStringLength int
}

// NewTruncatingSink returns a sink that truncates strings longer than the max
// data string length
// Example:
//   writerSink := lager.NewWriterSink(os.Stdout, lager.INFO)
//   sink := lager.NewTruncatingSink(testSink, 20)
//   logger := lager.NewLogger("test")
//   logger.RegisterSink(sink)
//   logger.Info("message", lager.Data{"A": strings.Repeat("a", 25)})
func NewTruncatingSink(sink Sink, maxDataStringLength int) Sink {
	return &truncatingSink{
		sink:                sink,
		maxDataStringLength: maxDataStringLength,
	}
}

func (sink *truncatingSink) Log(log LogFormat) {
	truncatedData := Data{}
	for k, v := range log.Data {
		truncatedData[k] = truncate.Value(v, sink.maxDataStringLength)
	}
	log.Data = truncatedData
	sink.sink.Log(log)
}
