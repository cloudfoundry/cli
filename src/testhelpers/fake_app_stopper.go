package testhelpers

import (
	"cf"
)

type FakeAppStopper struct {
	AppToStop cf.Application
	StoppedApp cf.Application
}

func (stopper *FakeAppStopper) ApplicationStop(app cf.Application) (updatedApp cf.Application, err error) {
	stopper.AppToStop = app
	updatedApp = stopper.StoppedApp
	return
}
