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
}

func (l *FakeLogsRepository) RecentLogsFor(app cf.Application, onConnect func(), onMessage func(*logmessage.Message), port string) (err error){
	l.logsFor(app, l.RecentLogs, onConnect, onMessage)
	return
}


func (l *FakeLogsRepository) TailLogsFor(app cf.Application, onConnect func(), onMessage func(*logmessage.Message),  printInterval time.Duration, port string) (err error){
	l.logsFor(app, l.TailLogMessages, onConnect, onMessage)
	return
}

func (l *FakeLogsRepository) logsFor(app cf.Application, logMessages []logmessage.LogMessage, onConnect func(), onMessage func(*logmessage.Message)) {
	l.AppLogged = app
	onConnect()
	for _, logMsg := range logMessages{
		data, _ := proto.Marshal(&logMsg)
		msg, _ := logmessage.ParseMessage(data)
		onMessage(msg)
	}

	return
}
