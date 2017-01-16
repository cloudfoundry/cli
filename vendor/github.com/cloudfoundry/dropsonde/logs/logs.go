// Package logs provides a simple API for sending app logs from STDOUT and STDERR
// through the dropsonde system.
//
// Use
//
// See the documentation for package dropsonde for configuration details.
//
// Importing package dropsonde and initializing will initial this package.
// To send logs use
//
//		logs.SendAppLog(appID, message, sourceType, sourceInstance)
//
// for sending errors,
//
//		logs.SendAppErrorLog(appID, message, sourceType, sourceInstance)
package logs

import (
	"io"

	"github.com/cloudfoundry/dropsonde/log_sender"
	"github.com/cloudfoundry/sonde-go/events"
)

type LogSender interface {
	SendAppLog(appID, message, sourceType, sourceInstance string) error
	SendAppErrorLog(appID, message, sourceType, sourceInstance string) error
	ScanLogStream(appID, sourceType, sourceInstance string, reader io.Reader)
	ScanErrorLogStream(appID, sourceType, sourceInstance string, reader io.Reader)
	LogMessage(msg []byte, msgType events.LogMessage_MessageType) log_sender.LogChainer
}

var logSender LogSender

// Initialize prepares the logs package for use with the automatic Emitter
// from dropsonde.
func Initialize(ls LogSender) {
	logSender = ls
}

// SendAppLog sends a log message with the given appid, log message, source type
// and source instance, with a message type of std out.
// Returns an error if one occurs while sending the event.
func SendAppLog(appID, message, sourceType, sourceInstance string) error {
	if logSender == nil {
		return nil
	}
	return logSender.SendAppLog(appID, message, sourceType, sourceInstance)
}

// SendAppErrorLog sends a log error message with the given appid, log message, source type
// and source instance, with a message type of std err.
// Returns an error if one occurs while sending the event.
func SendAppErrorLog(appID, message, sourceType, sourceInstance string) error {
	if logSender == nil {
		return nil
	}
	return logSender.SendAppErrorLog(appID, message, sourceType, sourceInstance)
}

// ScanLogStream sends a log message with the given meta-data for each line from reader.
// Restarts on read errors and continues until EOF.
func ScanLogStream(appID, sourceType, sourceInstance string, reader io.Reader) {
	if logSender == nil {
		return
	}
	logSender.ScanLogStream(appID, sourceType, sourceInstance, reader)
}

// ScanErrorLogStream sends a log error message with the given meta-data for each line from reader.
// Restarts on read errors and continues until EOF.
func ScanErrorLogStream(appID, sourceType, sourceInstance string, reader io.Reader) {
	if logSender == nil {
		return
	}
	logSender.ScanErrorLogStream(appID, sourceType, sourceInstance, reader)
}

// LogMessage creates a log message that can be manipulated via cascading calls
// and then sent.
func LogMessage(msg []byte, msgType events.LogMessage_MessageType) log_sender.LogChainer {
	return logSender.LogMessage(msg, msgType)
}
