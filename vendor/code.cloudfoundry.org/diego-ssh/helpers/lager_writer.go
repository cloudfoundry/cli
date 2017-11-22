package helpers

import (
	"io"

	"code.cloudfoundry.org/lager"
)

type lagerWriter struct {
	logger   lager.Logger
	logLevel lager.LogLevel
}

// NewLagerWriter wraps a Writer around a lager.Logger
// Log messages will be written at the specified log level
func NewLagerWriter(logger lager.Logger) io.Writer {
	return &lagerWriter{logger: logger}
}

func (lw *lagerWriter) Write(p []byte) (int, error) {
	lw.logger.Info("write", lager.Data{"payload": string(p)})
	return len(p), nil
}
