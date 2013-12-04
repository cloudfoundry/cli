package commands

import (
	"cf"
)

type FakeAppDisplayer struct {
	AppToDisplay cf.Application
}

func (displayer *FakeAppDisplayer) ShowApp(app cf.Application) {
	displayer.AppToDisplay = app
}
