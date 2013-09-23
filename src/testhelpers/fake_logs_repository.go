package testhelpers

import (
	"cf"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"time"
)

type FakeLogsRepository struct {
	AppLogged cf.Application
	RecentLogs []logmessage.LogMessage
	TailLogMessages []logmessage.LogMessage
}

func (l *FakeLogsRepository) RecentLogsFor(app cf.Application, onConnect func(), onMessage func(logmessage.LogMessage)) (err error){
	l.AppLogged = app
	onConnect()
	for _, message := range l.RecentLogs{
		onMessage(message)
	}

	return
}


func (l *FakeLogsRepository) TailLogsFor(app cf.Application, onConnect func(), onMessage func(logmessage.LogMessage),  printInterval time.Duration) (err error){
	l.AppLogged = app
	onConnect()
	for _, message := range l.TailLogMessages{
		onMessage(message)
	}

	return
}
