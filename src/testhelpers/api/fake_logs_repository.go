package api

import (
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"time"
)

type FakeLogsRepository struct {
	AppLoggedGuid string
	RecentLogs    []*logmessage.LogMessage
	RecentLogErr  error

	TailLogMessages []*logmessage.LogMessage
	TailLogErr      error

	TailLogStopCalled bool
}

func (l *FakeLogsRepository) RecentLogsFor(appGuid string) ([]*logmessage.LogMessage, error) {
	l.AppLoggedGuid = appGuid
	return l.RecentLogs, l.RecentLogErr
}

func (l *FakeLogsRepository) TailLogsFor(appGuid string, bufferTime time.Duration, onConnect func(), onMessage func(*logmessage.LogMessage)) (err error) {
	l.AppLoggedGuid = appGuid

	err = l.TailLogErr
	if err != nil {
		return
	}

	onConnect()

	for _, msg := range l.TailLogMessages {
		onMessage(msg)
	}

	return
}

func (l *FakeLogsRepository) Close() {
	l.TailLogStopCalled = true
}

func (l *FakeLogsRepository) logsFor(appGuid string, logMessages []*logmessage.LogMessage, onConnect func(), logChan chan *logmessage.LogMessage, stopLoggingChan chan bool) {
	l.AppLoggedGuid = appGuid
	onConnect()

	for _, logMsg := range logMessages {
		logChan <- logMsg
	}

	go func() {
		l.TailLogStopCalled = <-stopLoggingChan
	}()

	return
}
