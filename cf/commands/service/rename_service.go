package service

import (
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
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

func (command *RenameService) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "rename-service",
		Description: "Rename a service instance",
		Usage:       "CF_NAME rename-service SERVICE_INSTANCE NEW_SERVICE_INSTANCE",
	}
}

func (cmd *RenameService) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("incorrect usage")
		cmd.ui.FailWithUsage(c, "rename-service")
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

	cmd.ui.Say("Renaming service %s to %s in org %s / space %s as %s...",
		terminal.EntityNameColor(serviceInstance.Name),
		terminal.EntityNameColor(newName),
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.SpaceFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)
	err := cmd.serviceRepo.RenameService(serviceInstance, newName)

	if err != nil {
		if httpError, ok := err.(errors.HttpError); ok && httpError.ErrorCode() == errors.SERVICE_INSTANCE_NAME_TAKEN {
			cmd.ui.Failed("%s\nTIP: Use '%s services' to view all services in this org and space.", httpError.Error(), cf.Name())
		} else {
			cmd.ui.Failed(err.Error())
		}
	}

	cmd.ui.Ok()
}
