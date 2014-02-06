package commands

import (
"cf/models"
)

type FakeAppRestarter struct {
	AppToRestart models.Application
}

func (restarter *FakeAppRestarter) ApplicationRestart(appToRestart models.Application) {
	restarter.AppToRestart = appToRestart
	return
}
