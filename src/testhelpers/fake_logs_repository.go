package testhelpers

import (
	"cf"
	"cf/api"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
)

type FakeLogsRepository struct {
	AppLogged cf.Application
	RecentLogs []*logmessage.LogMessage
	TailLogMessages []*logmessage.LogMessage
}

func (l *FakeLogsRepository) RecentLogsFor(app cf.Application) (logs []*logmessage.LogMessage, apiErr *api.ApiError){
	l.AppLogged = app
	return l.RecentLogs, nil
}


func (l *FakeLogsRepository) TailLogsFor(app cf.Application, onConnect func(), onMessage func(*logmessage.LogMessage)) (err error){
	l.AppLogged = app
	onConnect()
	for _, message := range l.TailLogMessages{
		onMessage(message)
	}

	return
}
