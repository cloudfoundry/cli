package service

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type DeleteService struct {
	ui                 terminal.UI
	config             *configuration.Configuration
	serviceRepo        api.ServiceRepository
	serviceInstanceReq requirements.ServiceInstanceRequirement
}

func NewDeleteService(ui terminal.UI, config *configuration.Configuration, serviceRepo api.ServiceRepository) (cmd *DeleteService) {
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

	cmd.ui.Say("Deleting service %s in org %s / space %s as %s...",
		terminal.EntityNameColor(serviceName),
		terminal.EntityNameColor(cmd.config.Organization.Name),
		terminal.EntityNameColor(cmd.config.Space.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	instance, apiResponse := cmd.serviceRepo.FindInstanceByName(serviceName)

	if apiResponse.IsError() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	if apiResponse.IsNotFound() {
		cmd.ui.Ok()
		cmd.ui.Warn("Service %s does not exist.", serviceName)
		return
	}

	apiResponse = cmd.serviceRepo.DeleteService(instance)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
}
