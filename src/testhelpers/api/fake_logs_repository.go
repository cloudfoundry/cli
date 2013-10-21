package api

import (
	"cf"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"time"
	"code.google.com/p/gogoprotobuf/proto"
	"errors"
)

type FakeLogsRepository struct {
	AppLogged cf.Application
	Error string

	RecentLogs []logmessage.LogMessage
	TailLogMessages []logmessage.LogMessage
}

func (l *FakeLogsRepository) RecentLogsFor(app cf.Application, onConnect func(), onMessage func(*logmessage.Message), onError func(error)) (err error){
	l.logsFor(app, l.RecentLogs, onConnect, onMessage, onError)
	return
}


func (l *FakeLogsRepository) TailLogsFor(app cf.Application, onConnect func(), onMessage func(*logmessage.Message), onError func(error),  printInterval time.Duration) (err error){
	l.logsFor(app, l.TailLogMessages, onConnect, onMessage, onError)
	return
}

func (l *FakeLogsRepository) logsFor(app cf.Application, logMessages []logmessage.LogMessage, onConnect func(), onMessage func(*logmessage.Message), onError func(error)) {
	l.AppLogged = app
	onConnect()
	for _, logMsg := range logMessages{
		data, _ := proto.Marshal(&logMsg)
		msg, _ := logmessage.ParseMessage(data)
		onMessage(msg)
	}

	if l.Error != "" {
		onError(errors.New(l.Error))
	}

	return
}
