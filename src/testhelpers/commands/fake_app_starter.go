package commands

import (
	"cf"
)

type FakeAppStarter struct {
	AppToStart cf.Application
	Timeout    int
}

func (starter *FakeAppStarter) ApplicationStart(appToStart cf.Application) (startedApp cf.Application, err error) {
	starter.AppToStart = appToStart
	startedApp = appToStart
	return
}

func (starter *FakeAppStarter) SetStartTimeoutSeconds(timeout int) {
	starter.Timeout = timeout
}

func (starter *FakeAppStarter) ApplicationStartWithBuildpack(app cf.Application, buildpackUrl string) (startedApp cf.Application, err error) {
	starter.AppToStart = app
	startedApp = app
	return
}
