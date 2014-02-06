package commands

import (
"cf/models"
)

type FakeAppDisplayer struct {
	AppToDisplay models.Application
}

func (displayer *FakeAppDisplayer) ShowApp(app models.Application) {
	displayer.AppToDisplay = app
}
