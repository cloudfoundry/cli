package commands

import (
	"cf"
)

type FakeAppStarter struct {
	AppToStart cf.Application
}

func (starter *FakeAppStarter) ApplicationStart(appToStart cf.Application) (startedApp cf.Application, err error) {
	starter.AppToStart = appToStart
	startedApp = appToStart
	return
}

func (starter *FakeAppStarter) ApplicationStartWithBuildpack(app cf.Application, buildpackUrl string) (startedApp cf.Application, err error){
	starter.AppToStart = app
	startedApp = app
	return
}
