package testhelpers

import (
	"cf"
)

type FakeAppStarter struct {
	StartedApp cf.Application
}

func (starter *FakeAppStarter) ApplicationStart(app cf.Application) {
	starter.StartedApp = app
}
