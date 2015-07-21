package service

import (
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type RenameService struct {
	ui                 terminal.UI
	config             core_config.Reader
	serviceRepo        api.ServiceRepository
	serviceInstanceReq requirements.ServiceInstanceRequirement
}

func init() {
	command_registry.Register(&RenameService{})
}

func (cmd *RenameService) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "rename-service",
		Description: T("Rename a service instance"),
		Usage:       T("CF_NAME rename-service SERVICE_INSTANCE NEW_SERVICE_INSTANCE"),
	}
}

func (cmd *RenameService) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SERVICE_INSTANCE and NEW_SERVICE_INSTANCE as arguments\n\n") + command_registry.Commands.CommandUsage("rename-service"))
	}

	cmd.serviceInstanceReq = requirementsFactory.NewServiceInstanceRequirement(fc.Args()[0])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.serviceInstanceReq,
	}

	return
}

func (cmd *RenameService) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.serviceRepo = deps.RepoLocator.GetServiceRepository()
	return cmd
}

func (cmd *RenameService) Execute(c flags.FlagContext) {
	newName := c.Args()[1]
	serviceInstance := cmd.serviceInstanceReq.GetServiceInstance()

	cmd.ui.Say(T("Renaming service {{.ServiceName}} to {{.NewServiceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"ServiceName":    terminal.EntityNameColor(serviceInstance.Name),
			"NewServiceName": terminal.EntityNameColor(newName),
			"OrgName":        terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":      terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser":    terminal.EntityNameColor(cmd.config.Username()),
		}))
	err := cmd.serviceRepo.RenameService(serviceInstance, newName)

	if err != nil {
		if httpError, ok := err.(errors.HttpError); ok && httpError.ErrorCode() == errors.SERVICE_INSTANCE_NAME_TAKEN {
			cmd.ui.Failed(T("{{.ErrorDescription}}\nTIP: Use '{{.CFServicesCommand}}' to view all services in this org and space.",
				map[string]interface{}{
					"ErrorDescription":  httpError.Error(),
					"CFServicesCommand": cf.Name() + " " + "services",
				}))
		} else {
			cmd.ui.Failed(err.Error())
		}
	}

	cmd.ui.Ok()
}
