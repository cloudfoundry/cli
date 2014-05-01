package commands

import (
	"github.com/cloudfoundry/cli/cf/models"
)

type FakeAppStopper struct {
	AppToStop models.Application
}

func (stopper *FakeAppStopper) ApplicationStop(app models.Application) (updatedApp models.Application, err error) {
	stopper.AppToStop = app
	updatedApp = app
	return
}
