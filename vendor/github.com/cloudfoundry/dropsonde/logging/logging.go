package logging

import "github.com/cloudfoundry/gosteno"

// Debugf is a helper to avoid logging anything if the log level is not debug.
// It should be scrapped if/when we switch logging libraries to a library that
// doesn't do any processing if the log won't be output.
func Debugf(logger *gosteno.Logger, msg string, inputs ...interface{}) {
	switch logger.Level() {
	case gosteno.LOG_DEBUG, gosteno.LOG_DEBUG1, gosteno.LOG_DEBUG2:
		logger.Debugf(msg, inputs...)
	}
}
