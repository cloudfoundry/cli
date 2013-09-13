package testhelpers

import (
	"cf"
	"cf/api"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
)

type FakeLogsRepository struct {
	AppLogged cf.Application
	RecentLogs []*logmessage.LogMessage
}

func (l *FakeLogsRepository) RecentLogsFor(app cf.Application) (logs []*logmessage.LogMessage, apiErr *api.ApiError){
	l.AppLogged = app
	return l.RecentLogs, nil
}
