package servicebroker

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type RenameServiceBroker struct {
	ui     terminal.UI
	config coreconfig.Reader
	repo   api.ServiceBrokerRepository
}

func init() {
	commandregistry.Register(&RenameServiceBroker{})
}

func (cmd *RenameServiceBroker) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "rename-service-broker",
		Description: T("Rename a service broker"),
		Usage: []string{
			T("CF_NAME rename-service-broker SERVICE_BROKER NEW_SERVICE_BROKER"),
		},
	}
}

func (cmd *RenameServiceBroker) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SERVICE_BROKER, NEW_SERVICE_BROKER as arguments\n\n") + commandregistry.Commands.CommandUsage("rename-service-broker"))
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	return reqs
}

func (cmd *RenameServiceBroker) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.repo = deps.RepoLocator.GetServiceBrokerRepository()
	return cmd
}

func (cmd *RenameServiceBroker) Execute(c flags.FlagContext) error {
	serviceBroker, err := cmd.repo.FindByName(c.Args()[0])
	if err != nil {
		return err
	}

	cmd.ui.Say(T("Renaming service broker {{.OldName}} to {{.NewName}} as {{.Username}}",
		map[string]interface{}{
			"OldName":  terminal.EntityNameColor(serviceBroker.Name),
			"NewName":  terminal.EntityNameColor(c.Args()[1]),
			"Username": terminal.EntityNameColor(cmd.config.Username())}))

	newName := c.Args()[1]

	err = cmd.repo.Rename(serviceBroker.GUID, newName)

	if err != nil {
		return err
	}

	cmd.ui.Ok()
	return nil
}
