package testhelpers

import (
	"cf"
)

type FakeAppStopper struct {
	StoppedApp cf.Application
}

func (stopper *FakeAppStopper) ApplicationStop(app cf.Application) {
	stopper.StoppedApp = app
}
