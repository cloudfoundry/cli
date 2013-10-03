package service

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type DeleteService struct {
	ui                 terminal.UI
	serviceRepo        api.ServiceRepository
	serviceInstanceReq requirements.ServiceInstanceRequirement
}

func NewDeleteService(ui terminal.UI, sR api.ServiceRepository) (cmd *DeleteService) {
	cmd = new(DeleteService)
	cmd.ui = ui
	cmd.serviceRepo = sR
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

	cmd.ui.Say("Deleting service %s...", terminal.EntityNameColor(serviceName))

	instance, apiStatus := cmd.serviceRepo.FindInstanceByName(serviceName)

	if apiStatus.IsError() {
		cmd.ui.Failed(apiStatus.Message)
	}

	if apiStatus.IsNotFound() {
		cmd.ui.Ok()
		cmd.ui.Warn("Service %s does not exist.", serviceName)
		return
	}

	apiStatus = cmd.serviceRepo.DeleteService(instance)
	if apiStatus.NotSuccessful() {
		cmd.ui.Failed(apiStatus.Message)
		return
	}

	cmd.ui.Ok()
}
