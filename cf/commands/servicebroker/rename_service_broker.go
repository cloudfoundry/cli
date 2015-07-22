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

type RenameServiceBroker struct {
	ui     terminal.UI
	config core_config.Reader
	repo   api.ServiceBrokerRepository
}

func init() {
	command_registry.Register(&RenameServiceBroker{})
}

func (cmd *RenameServiceBroker) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "rename-service-broker",
		Description: T("Rename a service broker"),
		Usage:       T("CF_NAME rename-service-broker SERVICE_BROKER NEW_SERVICE_BROKER"),
	}
}

func (cmd *RenameServiceBroker) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SERVICE_BROKER, NEW_SERVICE_BROKER as arguments\n\n") + command_registry.Commands.CommandUsage("rename-service-broker"))
	}

	reqs = append(reqs, requirementsFactory.NewLoginRequirement())
	return
}

func (cmd *RenameServiceBroker) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.repo = deps.RepoLocator.GetServiceBrokerRepository()
	return cmd
}

func (cmd *RenameServiceBroker) Execute(c flags.FlagContext) {
	serviceBroker, apiErr := cmd.repo.FindByName(c.Args()[0])
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Say(T("Renaming service broker {{.OldName}} to {{.NewName}} as {{.Username}}",
		map[string]interface{}{
			"OldName":  terminal.EntityNameColor(serviceBroker.Name),
			"NewName":  terminal.EntityNameColor(c.Args()[1]),
			"Username": terminal.EntityNameColor(cmd.config.Username())}))

	newName := c.Args()[1]

	apiErr = cmd.repo.Rename(serviceBroker.Guid, newName)

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
