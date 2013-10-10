package commands

import (
	"cf"
)

type FakeAppRestarter struct {
	AppToRestart cf.Application
}

func (restarter *FakeAppRestarter) ApplicationRestart(appToRestart cf.Application) {
	restarter.AppToRestart = appToRestart
	return
}
