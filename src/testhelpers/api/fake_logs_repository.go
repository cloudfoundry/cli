package api

import (
	"cf"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"time"
	"code.google.com/p/gogoprotobuf/proto"
)

type FakeLogsRepository struct {
	AppLogged cf.Application
	RecentLogs []logmessage.LogMessage
	TailLogMessages []logmessage.LogMessage
	TailLogStopCalled bool
	TailLogErr error
}

func (l *FakeLogsRepository) RecentLogsFor(app cf.Application, onConnect func(), logChan chan *logmessage.Message) (err error){
	stopLoggingChan := make(chan bool)
	l.logsFor(app, l.RecentLogs, onConnect, logChan, stopLoggingChan)
	return
}


func (l *FakeLogsRepository) TailLogsFor(app cf.Application, onConnect func(), logChan chan *logmessage.Message, stopLoggingChan chan bool,  printInterval time.Duration) (err error){
	err = l.TailLogErr
	if err != nil {
		return
	}

	l.logsFor(app, l.TailLogMessages, onConnect, logChan, stopLoggingChan)
	return
}

func (l *FakeLogsRepository) logsFor(app cf.Application, logMessages []logmessage.LogMessage, onConnect func(), logChan chan *logmessage.Message, stopLoggingChan chan bool) {
	l.AppLogged = app
	onConnect()
	for _, logMsg := range logMessages{
		data, _ := proto.Marshal(&logMsg)
		msg, _ := logmessage.ParseMessage(data)
		logChan <- msg
	}

	close(logChan)

	go func(){
		l.TailLogStopCalled = <- stopLoggingChan
	}()

	return
}
