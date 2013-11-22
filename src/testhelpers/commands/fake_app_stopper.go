package commands

import (
	"cf"
)

type FakeAppStopper struct {
	AppToStop cf.Application
}

func (stopper *FakeAppStopper) ApplicationStop(app cf.Application) (updatedApp cf.Application, err error) {
	stopper.AppToStop = app
	updatedApp = app
	return
}
