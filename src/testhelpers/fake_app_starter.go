package testhelpers

import (
	"cf"
)

type FakeAppStarter struct {
	AppToStart cf.Application
	StartedApp cf.Application
}

func (starter *FakeAppStarter) ApplicationStart(appToStart cf.Application) (startedApp cf.Application, err error) {
	starter.AppToStart = appToStart
	startedApp = starter.StartedApp
	return
}
