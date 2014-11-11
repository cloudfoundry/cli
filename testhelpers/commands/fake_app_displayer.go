package commands

import (
	"github.com/cloudfoundry/cli/cf/models"
)

type FakeAppDisplayer struct {
	AppToDisplay models.Application
	OrgName      string
	SpaceName    string
}

func (displayer *FakeAppDisplayer) ShowApp(app models.Application, orgName, spaceName string) {
	displayer.AppToDisplay = app
}
