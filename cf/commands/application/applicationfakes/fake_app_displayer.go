package applicationfakes

import (
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/flags"
)

type FakeAppDisplayer struct {
	AppToDisplay models.Application
	OrgName      string
	SpaceName    string
}

func (displayer *FakeAppDisplayer) ShowApp(app models.Application, orgName, spaceName string) {
	displayer.AppToDisplay = app
}

func (displayer *FakeAppDisplayer) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{Name: "app"}
}

func (displayer *FakeAppDisplayer) SetDependency(_ commandregistry.Dependency, _ bool) commandregistry.Command {
	return displayer
}

func (displayer *FakeAppDisplayer) Requirements(_ requirements.Factory, _ flags.FlagContext) []requirements.Requirement {
	return []requirements.Requirement{}
}

func (displayer *FakeAppDisplayer) Execute(_ flags.FlagContext) {}
