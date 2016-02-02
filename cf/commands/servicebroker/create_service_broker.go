package servicebroker

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type CreateServiceBroker struct {
	ui                terminal.UI
	config            core_config.Reader
	serviceBrokerRepo api.ServiceBrokerRepository
}

func init() {
	command_registry.Register(&CreateServiceBroker{})
}

func (cmd *CreateServiceBroker) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "create-service-broker",
		Description: T("Create a service broker"),
		Usage:       T("CF_NAME create-service-broker SERVICE_BROKER USERNAME PASSWORD URL"),
	}
}

func (cmd *CreateServiceBroker) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 4 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SERVICE_BROKER, USERNAME, PASSWORD, URL as arguments\n\n") + command_registry.Commands.CommandUsage("create-service-broker"))
	}

	return []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}, nil
}

func (cmd *CreateServiceBroker) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.serviceBrokerRepo = deps.RepoLocator.GetServiceBrokerRepository()
	return cmd
}

func (cmd *CreateServiceBroker) Execute(c flags.FlagContext) {
	name := c.Args()[0]
	username := c.Args()[1]
	password := c.Args()[2]
	url := c.Args()[3]

	cmd.ui.Say(T("Creating service broker {{.Name}} as {{.Username}}...",
		map[string]interface{}{
			"Name":     terminal.EntityNameColor(name),
			"Username": terminal.EntityNameColor(cmd.config.Username())}))

	err := cmd.serviceBrokerRepo.Create(name, url, username, password)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}
