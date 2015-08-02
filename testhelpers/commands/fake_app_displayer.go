package commands

import (
	"github.com/cloudfoundry/cli/cf/command_registry"
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

func (cmd *FakeAppDisplayer) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{Name: "app"}
}

func (cmd *FakeAppDisplayer) SetDependency(_ command_registry.Dependency, _ bool) command_registry.Command {
	return cmd
}

func (cmd *FakeAppDisplayer) Requirements(_ requirements.Factory, _ flags.FlagContext) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd *FakeAppDisplayer) Execute(_ flags.FlagContext) {}
