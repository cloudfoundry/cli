package commands

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

	cmd.serviceInstanceReq = reqFactory.NewServiceInstanceRequirement(serviceName)

	reqs = []requirements.Requirement{cmd.serviceInstanceReq}
	return
}

func (cmd *DeleteService) Run(c *cli.Context) {
	instance := cmd.serviceInstanceReq.GetServiceInstance()
	cmd.ui.Say("Deleting service %s...", terminal.EntityNameColor(instance.Name))

	err := cmd.serviceRepo.DeleteService(instance)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Ok()
}
