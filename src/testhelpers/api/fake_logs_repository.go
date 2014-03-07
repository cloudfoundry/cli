package api

import (
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"time"
)

type FakeLogsRepository struct {
	AppLoggedGuid     string
	RecentLogs        []*logmessage.Message
	TailLogMessages   []*logmessage.Message
	TailLogStopCalled bool
	TailLogErr        error
}

func (l *FakeLogsRepository) RecentLogsFor(appGuid string, onConnect func(), logChan chan *logmessage.Message) (err error) {
	stopLoggingChan := make(chan bool)
	defer close(stopLoggingChan)
	l.logsFor(appGuid, l.RecentLogs, onConnect, logChan, stopLoggingChan)
	return
}

func (l *FakeLogsRepository) TailLogsFor(appGuid string, onConnect func(), logChan chan *logmessage.Message, stopLoggingChan chan bool, printInterval time.Duration) (err error) {
	err = l.TailLogErr

	if err != nil {
		return
	}

	l.logsFor(appGuid, l.TailLogMessages, onConnect, logChan, stopLoggingChan)
	return
}

func (l *FakeLogsRepository) logsFor(appGuid string, logMessages []*logmessage.Message, onConnect func(), logChan chan *logmessage.Message, stopLoggingChan chan bool) {
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
