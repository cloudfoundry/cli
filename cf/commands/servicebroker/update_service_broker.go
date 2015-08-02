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

type UpdateServiceBroker struct {
	ui     terminal.UI
	config core_config.Reader
	repo   api.ServiceBrokerRepository
}

func init() {
	command_registry.Register(&UpdateServiceBroker{})
}

func (cmd *UpdateServiceBroker) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "update-service-broker",
		Description: T("Update a service broker"),
		Usage:       T("CF_NAME update-service-broker SERVICE_BROKER USERNAME PASSWORD URL"),
	}
}

func (cmd *UpdateServiceBroker) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 4 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SERVICE_BROKER, USERNAME, PASSWORD, URL as arguments\n\n") + command_registry.Commands.CommandUsage("update-service-broker"))
	}

	reqs = append(reqs, requirementsFactory.NewLoginRequirement())
	return
}

func (cmd *UpdateServiceBroker) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.repo = deps.RepoLocator.GetServiceBrokerRepository()
	return cmd
}

func (cmd *UpdateServiceBroker) Execute(c flags.FlagContext) {
	serviceBroker, apiErr := cmd.repo.FindByName(c.Args()[0])
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Say(T("Updating service broker {{.Name}} as {{.Username}}...",
		map[string]interface{}{
			"Name":     terminal.EntityNameColor(serviceBroker.Name),
			"Username": terminal.EntityNameColor(cmd.config.Username())}))

	serviceBroker.Username = c.Args()[1]
	serviceBroker.Password = c.Args()[2]
	serviceBroker.Url = c.Args()[3]

	apiErr = cmd.repo.Update(serviceBroker)

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
