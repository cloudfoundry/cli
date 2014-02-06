package commands

import (
"cf/models"
)

type FakeAppStarter struct {
	AppToStart models.Application
	Timeout    int
}

func (starter *FakeAppStarter) ApplicationStart(appToStart models.Application) (startedApp models.Application, err error) {
	starter.AppToStart = appToStart
	startedApp = appToStart
	return
}

func (starter *FakeAppStarter) SetStartTimeoutSeconds(timeout int) {
	starter.Timeout = timeout
}

func (starter *FakeAppStarter) ApplicationStartWithBuildpack(app models.Application, buildpackUrl string) (startedApp models.Application, err error) {
	starter.AppToStart = app
	startedApp = app
	return
}
