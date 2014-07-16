package service

import (
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type RenameService struct {
	ui                 terminal.UI
	config             configuration.Reader
	serviceRepo        api.ServiceRepository
	serviceInstanceReq requirements.ServiceInstanceRequirement
}

func NewRenameService(ui terminal.UI, config configuration.Reader, serviceRepo api.ServiceRepository) (cmd *RenameService) {
	cmd = new(RenameService)
	cmd.ui = ui
	cmd.config = config
	cmd.serviceRepo = serviceRepo
	return
}

func (cmd *RenameService) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "rename-service",
		Description: T("Rename a service instance"),
		Usage:       T("CF_NAME rename-service SERVICE_INSTANCE NEW_SERVICE_INSTANCE"),
	}
}

func (cmd *RenameService) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		cmd.ui.FailWithUsage(c)
		return
	}

	cmd.serviceInstanceReq = requirementsFactory.NewServiceInstanceRequirement(c.Args()[0])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.serviceInstanceReq,
	}

	return
}

func (cmd *RenameService) Run(c *cli.Context) {
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
