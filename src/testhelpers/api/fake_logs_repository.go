package api

import (
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"time"
	"code.google.com/p/gogoprotobuf/proto"
)

type FakeLogsRepository struct {
	AppLoggedGuid string
	RecentLogs []logmessage.LogMessage
	TailLogMessages []logmessage.LogMessage
	TailLogStopCalled bool
	TailLogErr error
}

func (l *FakeLogsRepository) RecentLogsFor(appGuid string, onConnect func(), logChan chan *logmessage.Message) (err error){
	stopLoggingChan := make(chan bool)
	defer close(stopLoggingChan)
	l.logsFor(appGuid, l.RecentLogs, onConnect, logChan, stopLoggingChan)
	return
}


func (l *FakeLogsRepository) TailLogsFor(appGuid string, onConnect func(), logChan chan *logmessage.Message, stopLoggingChan chan bool,  printInterval time.Duration) (err error){
	err = l.TailLogErr
	if err != nil {
		return
	}

	l.logsFor(appGuid, l.TailLogMessages, onConnect, logChan, stopLoggingChan)
	return
}

func (l *FakeLogsRepository) logsFor(appGuid string, logMessages []logmessage.LogMessage, onConnect func(), logChan chan *logmessage.Message, stopLoggingChan chan bool) {
	l.AppLoggedGuid = appGuid
	onConnect()

	for _, logMsg := range logMessages{
		data, _ := proto.Marshal(&logMsg)
		msg, _ := logmessage.ParseMessage(data)
		logChan <- msg
	}

	go func(){
		l.TailLogStopCalled = <- stopLoggingChan
	}()

	return
}
