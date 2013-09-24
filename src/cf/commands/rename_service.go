package commands

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type RenameService struct {
	ui                 terminal.UI
	serviceRepo        api.ServiceRepository
	serviceInstanceReq requirements.ServiceInstanceRequirement
}

func NewRenameService(ui terminal.UI, serviceRepo api.ServiceRepository) (cmd *RenameService) {
	cmd = new(RenameService)
	cmd.ui = ui
	cmd.serviceRepo = serviceRepo
	return
}

func (cmd *RenameService) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("incorrect usage")
		cmd.ui.FailWithUsage(c, "rename-service")
		return
	}

	cmd.serviceInstanceReq = reqFactory.NewServiceInstanceRequirement(c.Args()[0])

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewTargetedSpaceRequirement(),
		cmd.serviceInstanceReq,
	}

	return
}

func (cmd *RenameService) Run(c *cli.Context) {
	newName := c.Args()[1]
	serviceInstance := cmd.serviceInstanceReq.GetServiceInstance()

	cmd.ui.Say("Renaming service %s...", serviceInstance.Name)
	err := cmd.serviceRepo.RenameService(serviceInstance, newName)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}
	cmd.ui.Ok()
}
