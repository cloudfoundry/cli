package service

import (
	"cf/api"
	"cf/configuration"
	"cf/errors"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type DeleteService struct {
	ui                 terminal.UI
	config             configuration.Reader
	serviceRepo        api.ServiceRepository
	serviceInstanceReq requirements.ServiceInstanceRequirement
}

func NewDeleteService(ui terminal.UI, config configuration.Reader, serviceRepo api.ServiceRepository) (cmd *DeleteService) {
	cmd = new(DeleteService)
	cmd.ui = ui
	cmd.config = config
	cmd.serviceRepo = serviceRepo
	return
}

func (cmd *DeleteService) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	var serviceName string

	if len(c.Args()) == 1 {
		serviceName = c.Args()[0]
	}

	if serviceName == "" {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "delete-service")
		return
	}

	return
}

func (cmd *DeleteService) Run(c *cli.Context) {
	serviceName := c.Args()[0]
	force := c.Bool("f")

	if !force {
		answer := cmd.ui.Confirm("Are you sure you want to delete the service %s ?", terminal.EntityNameColor(serviceName))
		if !answer {
			return
		}
	}

	cmd.ui.Say("Deleting service %s in org %s / space %s as %s...",
		terminal.EntityNameColor(serviceName),
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.SpaceFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	instance, apiErr := cmd.serviceRepo.FindInstanceByName(serviceName)

	switch apiErr.(type) {
	case nil:
	case errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn("Service %s does not exist.", serviceName)
		return
	default:
		cmd.ui.Failed(apiErr.Error())
		return
	}

	apiErr = cmd.serviceRepo.DeleteService(instance)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
